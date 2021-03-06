package gotube

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
)

func ExtractQueryParam(videoURL string) (string, error) {
	var id string
	var err error

	isMatch, err := regexp.MatchString(`https://www\.youtube\.com/watch\?v=[\w-]+`, videoURL) // TODO need better regex pattern
	if err != nil {
		return id, err
	}

	if !isMatch {
		return id, fmt.Errorf("Invalid YouTube URL")
	}

	var reprURL *url.URL
	reprURL, err = url.Parse(videoURL)
	if err != nil {
		return id, err
	}
	id = reprURL.Query()["v"][0]

	return id, nil
}

func GetMetaData(url string) (Video, error) {
	id, err := ExtractQueryParam(url)
	if err != nil {
		return Video{}, fmt.Errorf("ExtractQueryParam failed")
	}

	resp, err := http.Get(fmt.Sprintf("https://www.youtube.com/watch?v=%s", id))
	if err != nil {
		return Video{}, fmt.Errorf("get video failed")
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Video{}, fmt.Errorf("failed parsing response body")
	}
	extractJSON := ExtractValue(string(bodyBytes), "ytInitialPlayerResponse = ", ";</script>")
	var youtubeRequest Video
	json.Unmarshal([]byte(extractJSON), &youtubeRequest)

	return youtubeRequest,nil
}