package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"start_build/internal"
	"strconv"

	"start_build/builtin"

	"start_build/service"

	"start_build/helper"

	buildimage "start_build/buildImage"

	remotbuild "start_build/remoteBuild"

	"net/http"

	"github.com/codeclysm/extract"
)

type StartBuildInput struct {
	AppId         string                 `json:"appId"`
	SourceUrl     string                 `json:"sourceUrl"`
	SourceType    string                 `json:"sourceType"`
	BuildType     string                 `josn:"buildType"`
	ImageTag      string                 `json:"imageTag"`
	BuildArgs     map[string]interface{} `json:"buildArgs"`
	FileExtension string                 `json:"fileExtension"`
	InternalPort  string                 `json:"internalPort"`
}

func StartBuild(w http.ResponseWriter, r *http.Request) {
	// w.Header().Set("Content-Type", "application/json")
	var input StartBuildInput
	json.NewDecoder(r.Body).Decode(&input)

	var ArchiveFile io.Reader
	var Out os.File

	_, err := internal.ValidateFileExtension(input.FileExtension)
	if err != nil {
		helper.RespondwithJSON(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
	}

	if input.FileExtension == "zip" || input.FileExtension == "tar" || input.FileExtension == "tar.gz" {
		out, err := os.Create(input.AppId)
		if err != nil {
			helper.RespondwithJSON(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
			return
		}
		defer out.Close()
		Out = *out
		reader, err := service.GetFileFromS3(input.SourceUrl)
		if err != nil {
			log.Println(err)
			helper.RespondwithJSON(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
			return
		}
		archiveFile := bytes.NewReader(reader)

		switch input.FileExtension {
		case "zip":
			extract.Zip(context.Background(), archiveFile, "extracted_file/"+input.AppId, nil)
		case "tar.gz":
			extract.Gz(context.Background(), archiveFile, "extracted_file/"+input.AppId, nil)
		case "tar":
			extract.Tar(context.Background(), archiveFile, "extracted_file/"+input.AppId, nil)
		}

		if input.BuildType == "Builtin" {

			filePath, _ := service.FindFile("extracted_file/" + input.AppId)

			if filePath == "Docker file doesn't exists" {
				builtInFile, _ := builtin.GetBuiltIn()
				internalPort, _ := strconv.Atoi(input.InternalPort)
				_, err = builtin.CreateDockerFile(int64(internalPort), builtInFile[input.SourceType], input.AppId)
				if err != nil {
					log.Println(err)
					Out.Close()
					err = service.DeletedSourceFile("extracted_file/" + input.AppId)
					if err != nil {
						helper.RespondwithJSON(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
						return
					}
				}
			}
		}

		fmt.Println("reached find file")
		filePath, err := service.FindFile("extracted_file/" + input.AppId)
		if filePath == "Docker file doesn't exists" {
			//Delete extracted file and image

			err = service.DeletedSourceFile("extracted_file/" + input.AppId)
			if err != nil {
				helper.RespondwithJSON(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
				return
			}
			helper.RespondwithJSON(w, http.StatusBadRequest, map[string]string{"message": filePath})
		}
		if err != nil {
			log.Println(err)
			Out.Close()
			err = service.DeletedSourceFile("extracted_file/" + input.AppId)
			if err != nil {
				helper.RespondwithJSON(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
				return
			}
			helper.RespondwithJSON(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
			return
		}

		fmt.Println("finished find file" + filePath)

		buildContext, err := remotbuild.NewBuildContext()
		if err != nil {
			helper.RespondwithJSON(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
			return
		}
		defer buildContext.Close()

		buildContext.AddSourceFolder(filePath, "")

		archive, err := buildContext.Archive()
		if err != nil {
			helper.RespondwithJSON(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
			return
		}
		defer archive.Close()

		ArchiveFile = archive

		source, sourceError := os.Open(archive.Name())
		if sourceError != nil {
			log.Println(sourceError)
		}
		defer source.Close()
		err = os.RemoveAll("extracted_file/" + input.AppId)
		if err != nil {
			log.Println(err)
		}

		ArchiveFile = source
	}

	if input.FileExtension == "file" {
		out, err := os.Create(input.AppId)
		if err != nil {
			helper.RespondwithJSON(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
			return
		}
		defer out.Close()
		Out = *out

		reader, err := service.GetFileFromS3(input.SourceUrl)
		if err != nil {
			log.Println(err)
			helper.RespondwithJSON(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
			return

		}
		ArchiveFile = bytes.NewReader(reader)
	}
	img, err := buildimage.BuildImage(context.TODO(), ArchiveFile, input.ImageTag, &Out, input.BuildArgs)

	if err != nil {
		Out.Close()
		_ = os.Remove(input.AppId)
		helper.RespondwithJSON(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
		return
	}
	_ = os.Remove(input.AppId)
	w.Write([]byte(img.Tag))
}
