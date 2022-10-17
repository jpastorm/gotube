package main

import (
	"fmt"
	"github.com/jpastorm/gotube"
	"log"
)

func main() {
	yr, err := gotube.GetMetaData("https://www.youtube.com/watch?v=AvQLyCqOyFs")
	if err != nil {
		fmt.Println(err)
	}

	downloadPath, err := gotube.Download(".", yr.StreamingData.Formats[0].URL, "myVideo")
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(downloadPath)
}
