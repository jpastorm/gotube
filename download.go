package gotube

import (
	"github.com/cavaliergopher/grab/v3"
	"time"
)

var downloadProgressList = make(map[string]float64)

// Download returns the size of a download
func Download(downloadPath, url string, name string) (string, error) {
	client := grab.NewClient()
	req, _ := grab.NewRequest(downloadPath, url)

	resp := client.Do(req)

	t := time.NewTicker(500 * time.Millisecond)
	defer t.Stop()

Loop:
	for {
		select {
		case <-t.C:
			downloadProgressList[name] = resp.Progress() * 100
		case <-resp.Done:
			break Loop
		}
	}

	// check for errors
	if err := resp.Err(); err != nil {
		return "", err
	}

	return downloadPath, nil
}

// GetDownloadProgressList returns a map with the current download percentage of the running processes
func GetDownloadProgressList() map[string]float64 {
	return downloadProgressList
}

// GetDownloadProgress returns an element with the current download percentage of the running processes
func GetDownloadProgress(name string) float64 {
	return downloadProgressList[name]
}

// ClearProgressList clear download in progress list
func ClearProgressList() {
	downloadProgressList = make(map[string]float64)
}

// DeleteElementFromProgressList Remove an item from the progress list
func DeleteElementFromProgressList(name string) {
	delete(downloadProgressList, name)
}
