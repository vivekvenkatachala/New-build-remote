package service

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
)

func GetFileFromS3(SourceURL string) ([]byte, error) {
	url := SourceURL
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	reader, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return reader, nil
}

func FindFile(path string) (string, error) {

	f, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	files, err := f.Readdir(0)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	var pathName string
	for _, v := range files {
		pathName = v.Name()
		if pathName == "__MACOSX" {
			continue
		} else {
			break
		}

	}

	filePath := path + "/" + pathName
	f, err = os.Open(filePath)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	files, err = f.Readdir(0)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	for _, v := range files {
		if string(v.Name()) == "Dockerfile" {
			filePath = path + "/" + pathName
			return filePath, err
		} else {
			filePath = ""
		}

	}

	if filePath == "" {
		filePath = "Docker file doesn't exists"
	}

	return filePath, err
}

func DeletedSourceFile(filePath string) error {
	err := os.RemoveAll(filePath)
	if err != nil {
		log.Println(err)
	}
	file := path.Base(filePath)
	err = os.Remove(file)
	if err != nil {
		log.Println(err)
	}
	return err
}
