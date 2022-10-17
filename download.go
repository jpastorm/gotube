package gotube

import (
	"time"
)

var downloadProgressList = make(map[string]float64)

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

func GetDownloadProgressList() map[string]float64 {
	return downloadProgressList
}

func GetDownloadProgress(name string) float64 {
	return downloadProgressList[name]
}

func ClearProgressList() {
	downloadProgressList = make(map[string]float64)
}

func DeleteElementFromProgressList(name string) {
	delete(downloadProgressList, name)
}
