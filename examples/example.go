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
	gotube.DownloadFile(yr.StreamingData.Formats[0].URL,"videos" ,yr.VideoDetails.Title+".mp4")
}