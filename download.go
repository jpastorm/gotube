package gotube

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"
)

func createDir(path string) error {
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}
func directoyExist(path string) error {
	var fileInfo os.FileInfo
	fileInfo, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("directory %s doesnt exist, '%v' ", path, err)
		}
		return err
	}
	if !fileInfo.Mode().IsDir() {
		return fmt.Errorf("The directory %s is a file, %v", path, err)
	}

	return nil
}

func DownloadFile(url, folder, filename string, showProgress bool)  error {
	if showProgress {
		return DownloadFileProgress(url, folder, filename)
	}
	return DownloadFileWithoutProgress(url, folder, filename)
}

func DownloadFileProgress(url, folder, fileName string) error {
	if err := directoyExist(folder); err != nil {
		if err = createDir(folder); err != nil {
			return err
		}
	}

	file := path.Base(url)

	log.Printf("Downloading file %s from %s\n", file, url)

	var path bytes.Buffer
	path.WriteString(folder)
	path.WriteString("/")
	path.WriteString(fileName)

	start := time.Now()

	out, err := os.Create(path.String())

	if err != nil {
		return err
	}

	defer out.Close()

	headResp, err := http.Head(url)

	if err != nil {
		return err
	}

	defer headResp.Body.Close()

	size, err := strconv.Atoi(headResp.Header.Get("Content-Length"))

	if err != nil {
		return err
	}

	done := make(chan int64)

	go PrintDownloadPercent(done, path.String(), int64(size))

	resp, err := http.Get(url)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	n, err := io.Copy(out, resp.Body)

	if err != nil {
		return err
	}

	done <- n

	elapsed := time.Since(start)
	log.Printf("Download completed in %s", elapsed)

	return nil
}

func DownloadFileInMemory(uri string) ([]byte, error) {
	res, err := http.Get(uri)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	d, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return d, err
}

func PrintDownloadPercent(done chan int64, path string, total int64) {
	var stop = false
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	for {
		select {
		case <-done:
			stop = true
		default:
			fi, err := file.Stat()
			if err != nil {
				log.Fatal(err)
			}

			size := fi.Size()
			if size == 0 {
				size = 1
			}

			var percent = float64(size) / float64(total) * 100
			fmt.Printf("%.0f", percent)
			fmt.Println("%")
		}

		if stop {
			break
		}
		time.Sleep(time.Second)
	}
}

func DownloadFileWithoutProgress(URL, path, fileName string) error {
	if err := directoyExist(path); err != nil {
		if err = createDir(path); err != nil {
			return err
		}
	}

	response, err := http.Get(URL)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return errors.New("Received non 200 response code")
	}

	file, err := os.Create(filepath.Join(path, fileName))
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}

	return nil
}