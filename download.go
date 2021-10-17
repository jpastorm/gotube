package gotube

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func createDir(path string) error {
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}
func directoyExist(path string) error {
	var fileInfo os.FileInfo
	fileInfo, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("directory %s doesnt exist, '%v' ", path, err)
		}
		return err
	}
	if !fileInfo.Mode().IsDir() {
		return fmt.Errorf("The directory %s is a file, %v", path, err)
	}

	return nil
}

func DownloadFile(URL, path, fileName string) error {
	if err := directoyExist(path); err != nil {
		if err = createDir(path); err != nil {
			return err
		}
	}

	response, err := http.Get(URL)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return errors.New("Received non 200 response code")
	}

	file, err := os.Create(filepath.Join(path, fileName))
	if err != nil {
		return err
	}
	defer file.Close()

	//Write the bytes to the fiel
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}

	return nil
}

func DownloadFileInMemory(uri string) ([]byte, error) {
	res, err := http.Get(uri)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	d, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	return d, err
}