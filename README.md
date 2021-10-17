<p align="center">
  A youtube library for retrieving metadata, and obtaining direct links to video-only/audio-only/mixed versions of videos on YouTube in Go.
</p>

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
)

func main() {
	yr, err := gotube.GetMetaData("https://www.youtube.com/watch?v=AvQLyCqOyFs")
	if err != nil {
		fmt.Println(err)
	}
	gotube.DownloadFile(yr.StreamingData.Formats[0].URL,"videos" ,yr.VideoDetails.Title+".mp4", true)
}
```