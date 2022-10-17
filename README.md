# Gotube

[![Go report](https://goreportcard.com/badge/github.com/jpastorm/gotube)](https://goreportcard.com/report/github.com/jpastorm/gotube)

<p align="center">
✨ A youtube library for retrieving metadata, obtaining direct links to video-only/audio-only/mixed versions of videos and download videos.
</p>

## 🤠 Docs

GoDoc: [https://pkg.go.dev/github.com/jpastorm/gotube](https://godoc.org/github.com/jpastorm/gotube)

## Install
```
go get github.com/jpastorm/gotube
```
## ⚡️ Quickstart

```go
package main

import (
	"fmt"
	"github.com/jpastorm/gotube"
	"log"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup

	wg.Add(3)
	fmt.Println("Start Goroutines")

	go download("./video1", "https://www.youtube.com/watch?v=AvQLyCqOyFs", "video1", &wg)
	go download("./video2", "https://www.youtube.com/watch?v=yOt1bL-s0n8", "video2", &wg)
	go download("./video3", "https://www.youtube.com/watch?v=P5DeAD_uXE0", "video3", &wg)
	go startTimer()
	// Wait for the goroutines to finish.
	wg.Wait()
	fmt.Println("Terminating Program")
}

func download(path, url, name string, group *sync.WaitGroup) {
	defer group.Done()
	yr, err := gotube.GetMetaData(url)
	if err != nil {
		fmt.Println(err)
	}

	downloadPath, err := gotube.Download(path, yr.StreamingData.Formats[0].URL, name)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(downloadPath)
}

func startTimer() {
	for {
		time.Sleep(2 * time.Second)
		fmt.Println(gotube.GetDownloadProgressList())
	}
}
```