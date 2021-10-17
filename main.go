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