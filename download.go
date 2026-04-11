package gotube

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

const (
	defaultChunks    = 16
	defaultBufSize   = 256 * 1024 // 256 KB per read
	minChunkSize     = 512 * 1024 // don't split below 512 KB
	progressInterval = 250 * time.Millisecond
)

var (
	progressMu       sync.RWMutex
	downloadProgress = make(map[string]float64)

	// shared optimised transport for all downloads
	sharedTransport = &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: false},
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 32,
		MaxConnsPerHost:     32,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 5 * time.Second,
		WriteBufferSize:     defaultBufSize,
		ReadBufferSize:      defaultBufSize,
		ForceAttemptHTTP2:   true,
	}

	sharedClient = &http.Client{
		Transport: sharedTransport,
		Timeout:   0, // no overall timeout; per-chunk timeouts are handled via context
	}
)

// Download downloads the file at url to downloadPath using parallel HTTP range
// requests and reports progress under name.
func Download(downloadPath, url string, name string) (string, error) {
	return DownloadWithContext(context.Background(), downloadPath, url, name)
}

// DownloadWithContext is like Download but accepts a context for cancellation.
func DownloadWithContext(ctx context.Context, downloadPath, url string, name string) (string, error) {
	return DownloadWithOptions(ctx, downloadPath, url, name, defaultChunks)
}

// DownloadWithOptions allows specifying the number of parallel chunks.
func DownloadWithOptions(ctx context.Context, downloadPath, url string, name string, numChunks int) (string, error) {
	if numChunks < 1 {
		numChunks = 1
	}

	// probe for total size & range support
	totalSize, supportsRange, err := probeURL(ctx, url)
	if err != nil {
		return "", fmt.Errorf("probe url: %w", err)
	}

	// fall back to single-stream when server doesn't support ranges or file is small
	if !supportsRange || totalSize <= 0 || totalSize < minChunkSize {
		return downloadSingle(ctx, downloadPath, url, name, totalSize)
	}

	return downloadParallel(ctx, downloadPath, url, name, totalSize, numChunks)
}

// probeURL sends a HEAD request to get Content-Length and check Accept-Ranges.
func probeURL(ctx context.Context, url string) (int64, bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return 0, false, err
	}
	resp, err := sharedClient.Do(req)
	if err != nil {
		return 0, false, err
	}
	resp.Body.Close()

	size := resp.ContentLength
	supportsRange := resp.Header.Get("Accept-Ranges") == "bytes"
	return size, supportsRange, nil
}

// downloadParallel splits the file in N chunks and downloads concurrently.
func downloadParallel(ctx context.Context, path, url, name string, totalSize int64, numChunks int) (string, error) {
	chunkSize := totalSize / int64(numChunks)
	if chunkSize < minChunkSize {
		numChunks = int(totalSize / minChunkSize)
		if numChunks < 1 {
			numChunks = 1
		}
		chunkSize = totalSize / int64(numChunks)
	}

	file, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}

	// pre-allocate file on disk so parallel writers don't collide
	if err := file.Truncate(totalSize); err != nil {
		file.Close()
		return "", fmt.Errorf("truncate: %w", err)
	}

	var downloaded int64
	setProgress(name, 0)

	// progress reporter goroutine
	done := make(chan struct{})
	go func() {
		t := time.NewTicker(progressInterval)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				d := atomic.LoadInt64(&downloaded)
				setProgress(name, float64(d)/float64(totalSize)*100)
			case <-done:
				return
			}
		}
	}()

	errCh := make(chan error, numChunks)
	var wg sync.WaitGroup

	for i := 0; i < numChunks; i++ {
		start := int64(i) * chunkSize
		end := start + chunkSize - 1
		if i == numChunks-1 {
			end = totalSize - 1 // last chunk gets remainder
		}

		wg.Add(1)
		go func(start, end int64) {
			defer wg.Done()
			if err := downloadChunk(ctx, file, url, start, end, &downloaded); err != nil {
				errCh <- err
			}
		}(start, end)
	}

	wg.Wait()
	close(done)
	close(errCh)

	file.Close()

	// collect first error
	for e := range errCh {
		if e != nil {
			os.Remove(path)
			return "", e
		}
	}

	setProgress(name, 100)
	return path, nil
}

// downloadChunk fetches a byte range and writes it at the correct offset.
func downloadChunk(ctx context.Context, file *os.File, url string, start, end int64, downloaded *int64) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))

	resp, err := sharedClient.Do(req)
	if err != nil {
		return fmt.Errorf("chunk request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d for range %d-%d", resp.StatusCode, start, end)
	}

	buf := make([]byte, defaultBufSize)
	offset := start

	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			_, wErr := file.WriteAt(buf[:n], offset)
			if wErr != nil {
				return fmt.Errorf("write at offset %d: %w", offset, wErr)
			}
			offset += int64(n)
			atomic.AddInt64(downloaded, int64(n))
		}
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return fmt.Errorf("read chunk: %w", readErr)
		}
	}
	return nil
}

// downloadSingle uses a single GET request (fallback when ranges not supported).
func downloadSingle(ctx context.Context, path, url, name string, totalSize int64) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	resp, err := sharedClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	file, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	setProgress(name, 0)

	buf := make([]byte, defaultBufSize)
	var downloaded int64

	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, wErr := file.Write(buf[:n]); wErr != nil {
				return "", wErr
			}
			downloaded += int64(n)
			if totalSize > 0 {
				setProgress(name, float64(downloaded)/float64(totalSize)*100)
			}
		}
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return "", readErr
		}
	}

	setProgress(name, 100)
	return path, nil
}

// --- progress tracking (thread-safe) ---

func setProgress(name string, pct float64) {
	progressMu.Lock()
	downloadProgress[name] = pct
	progressMu.Unlock()
}

// GetDownloadProgressList returns a snapshot of current download percentages.
func GetDownloadProgressList() map[string]float64 {
	progressMu.RLock()
	defer progressMu.RUnlock()
	cp := make(map[string]float64, len(downloadProgress))
	for k, v := range downloadProgress {
		cp[k] = v
	}
	return cp
}

// GetDownloadProgress returns the current download percentage for name.
func GetDownloadProgress(name string) float64 {
	progressMu.RLock()
	defer progressMu.RUnlock()
	return downloadProgress[name]
}

// ClearProgressList clears the download progress list.
func ClearProgressList() {
	progressMu.Lock()
	downloadProgress = make(map[string]float64)
	progressMu.Unlock()
}

// DeleteElementFromProgressList removes an item from the progress list.
func DeleteElementFromProgressList(name string) {
	progressMu.Lock()
	delete(downloadProgress, name)
	progressMu.Unlock()
}
