package gotube

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
)

// ExtractQueryParam Extract the id from a youtube url
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

// GetMetaData Get all the data of a youtube video and return it in json format
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
	extractJSON := extractValue(string(bodyBytes), "ytInitialPlayerResponse = ", ";</script>")
	var youtubeRequest Video
	err = json.Unmarshal([]byte(extractJSON), &youtubeRequest)
	if err != nil {
		return Video{}, fmt.Errorf("failed extract json")
	}

	return youtubeRequest, nil
}

// GetDownloadSize returns the size of a download
func GetDownloadSize(url string) (int64, error) {
	resp, err := http.Head(url)
	if err != nil {
		return 0, err
	}

	size, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	if err != nil {
		return 0, err
	}

	downloadSize := int64(size)

	return downloadSize, nil
}
