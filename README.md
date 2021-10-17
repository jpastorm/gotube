<p align="center">
  A youtube library for retrieving metadata, and obtaining direct links to video-only/audio-only/mixed versions of videos on YouTube in Go.
</p>

## ⚡️ Quickstart

```go
package main

import (
	"fmt"
)

func main() {
	id, err := ExtractQueryParam("https://www.youtube.com/watch?v=AvQLyCqOyFs")
	if err != nil {
		fmt.Println(err)
	}
	yr := GetMetaData(id)
	DownloadFile(yr.StreamingData.Formats[0].URL,"videos" ,yr.VideoDetails.Title+".mp4")
}
```