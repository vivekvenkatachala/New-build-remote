package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"start_build/internal"
	"strconv"

	"start_build/builtin"

	"start_build/service"

	"start_build/helper"

	buildimage "start_build/buildImage"

	remotbuild "start_build/remoteBuild"

	"net/http"

	"github.com/alecthomas/log4go"
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
	DockerFile    string                 `json:"dockerFile"`
}

func StartBuild(w http.ResponseWriter, r *http.Request) {
	// w.Header().Set("Content-Type", "application/json")
	var input StartBuildInput
	json.NewDecoder(r.Body).Decode(&input)

	var ArchiveFile io.Reader
	var Out os.File

	_, err := internal.ValidateFileExtension(input.FileExtension)
	if err != nil {
		log4go.Error("Module: StartBuild, MethodName: ValidateFileExtension, Message: %s ", err.Error())
		helper.RespondwithJSON(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
		return
	}
	log4go.Info("Module: StartBuild, MethodName: ValidateFileExtension, Message: validating the file extension. The file extension is %s", input.FileExtension)

	if input.FileExtension == "zip" || input.FileExtension == "tar" || input.FileExtension == "tar.gz" {
		out, err := os.Create(input.AppId)
		if err != nil {
			log4go.Error("Module: StartBuild, MethodName: Create, Message: %s ", err.Error())
			helper.RespondwithJSON(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
			return
		}
		defer out.Close()
		Out = *out
		reader, err := service.GetFileFromS3(input.SourceUrl)
		if err != nil {
			log.Println(err)
			log4go.Error("Module: StartBuild, MethodName: GetFileFromS3, Message: %s ", err.Error())
			helper.RespondwithJSON(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
			return
		}
		log4go.Info("Module: StartBuild, MethodName: GetFileFromS3, Message: Pulled the %s file from the S3 with the source url - %s ", input.FileExtension, input.SourceUrl)

		archiveFile := bytes.NewReader(reader)

		switch input.FileExtension {
		case "zip":
			extract.Zip(context.Background(), archiveFile, "extracted_file/"+input.AppId, nil)
		case "tar.gz":
			extract.Gz(context.Background(), archiveFile, "extracted_file/"+input.AppId, nil)
		case "tar":
			extract.Tar(context.Background(), archiveFile, "extracted_file/"+input.AppId, nil)
		}
		log4go.Info("Module: StartBuild, MethodName: GetFileFromS3, Message: The %s file is Extracted to the path extracted_file/ %s ", input.FileExtension, input.AppId)
		//The input.fileextnsion file is Extracted to the path "extracted_file/"+input.AppId
		// ---------------------------------------------------------------------------------------
		var fileName string
		files, err := ioutil.ReadDir("extracted_file/" + input.AppId)
		if err != nil {
			return
		}
		for _, i := range files {
			fileName = i.Name()
		}

		path := filepath.Join("extracted_file/", input.AppId, fileName, "sonarscan.sh")
		_, err = os.Create(path)
		if err != nil {
			log.Fatal(err)
		}
		data, err := ioutil.ReadFile("dotnet.sh")
		if err != nil {
			return
		}
		data1 := []byte(data)

		err = ioutil.WriteFile(path, data1, 0644)
		if err != nil {
			return
		}

		_, err = exec.Command("bash", "extracted_file/"+input.AppId+"/WebApplication4/sonarscan.sh").Output()
		if err != nil {
			fmt.Println(err)
			return
		}

		// ---------------------------------------------------------------------------------------

		if input.BuildType == "Builtin" {

			fmt.Println("inside builtin")

			filePath, _ := service.FindFile("extracted_file/" + input.AppId)
			log4go.Info("Module: StartBuild, MethodName: FindFile, Message: Searching for docker file in the Extracted file - extracted_file/%s ", input.AppId)
			//Searching for docker file in the Extracted file - extracted_file/" + input.AppId

			if filePath == "Docker file doesn't exists" {
				log4go.Info("Module: StartBuild, Message: Docker file doesn't exists in the Extracted file - extracted_file/%s ", input.AppId)
				//Docker file doesn't exists in the Extracted file - extracted_file/" + input.AppId

				builtInFile, _ := builtin.GetBuiltIn()
				log4go.Info("Module: StartBuild, MethodName: GetBuiltIn, Message: Docker file doesn't exists in the file. So, Fetching all the available Builtins ")
				//Fetching all the default Builtins

				internalPort, _ := strconv.Atoi(input.InternalPort)

				if input.DockerFile != "" {
					log4go.Info("Module: StartBuild, Message: User added the Dockerfile on the fly, Dockerfile: ", input.DockerFile)
					getPath, err := builtin.FindPath("extracted_file/" + input.AppId)

					if err != nil {
						log4go.Error("Module: StartBuild, MethodName: FindPath, Message: %s ", err.Error())
						log.Println(err)
						return
					}
					if !service.FileExists("extracted_file/" + input.AppId + "/" + getPath + "/Dockerfile") {
						log4go.Info("Module: StartBuild, MethodName: FileExists, Message: Searching for Dockerfile in the extracted_file/" + input.AppId + "/" + getPath + "/Dockerfile" + ". To re-write the dockerfile with the updated dockerfile, which is added on the fly")
						_, err := os.Create("extracted_file/" + input.AppId + "/" + getPath + "/Dockerfile")
						if err != nil {
							log4go.Error("Module: StartBuild, MethodName: Create, Message: %s ", err.Error())
							return
						}
						log4go.Info("Module: StartBuild, MethodName: Create, Message: Docker file doesn't exists in the Extracted file - " + input.AppId + "/" + getPath + "/Dockerfile" + ". Creating a Docker file to add the dockerfile, which is added on the fly")
					}
					data := []byte(input.DockerFile)

					err = ioutil.WriteFile("extracted_file/"+input.AppId+"/"+getPath+"/Dockerfile", data, 0)
					if err != nil {
						log4go.Error("Module: StartBuild, MethodName: WriteFile, Message: %s ", err.Error())
						log.Fatal(err)
						err = service.DeletedSourceFile("extracted_file/" + input.AppId)
						if err != nil {
							log4go.Error("Module: StartBuild, MethodName: DeletedSourceFile, Message: %s ", err.Error())
							helper.RespondwithJSON(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
							return
						}
						log4go.Info("Module: StartBuild, MethodName: DeletedSourceFile, Message: Error occurred while writing the dockerfile. So, Deleting the Source file from the path : extracted_file/" + input.AppId)
					}
					log4go.Info("Module: StartBuild, MethodName: WriteFile, Message: Adding the Docker file content to the dockerfile, which is added on the fly. Dockerfile: " + input.DockerFile)
				} else {
					_, err = builtin.CreateDockerFile(int64(internalPort), builtInFile[input.SourceType], input.AppId)
					if err != nil {
						log4go.Error("Module: StartBuild, MethodName: CreateDockerFile, Message: %s ", err.Error())
						log.Println(err)
						Out.Close()
						err = service.DeletedSourceFile("extracted_file/" + input.AppId)
						if err != nil {
							log4go.Error("Module: StartBuild, MethodName: DeletedSourceFile, Message: %s ", err.Error())
							helper.RespondwithJSON(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
							return
						}
						log4go.Info("Module: StartBuild, MethodName: DeletedSourceFile, Message: Error occurred while creating dockerfile using default builtins. So, Deleting the extracted source file")
					}
					log4go.Info("Module: StartBuild, MethodName: CreateDockerFile, Message: Docker file doesn't exists in the Extracted file. So, Adding the dockfile from the default available builtins")
				}
			}
		}

		fmt.Println("reached find file")

		if input.DockerFile != "" {
			log4go.Info("Module: StartBuild, Message: User added the Dockerfile on the fly, Dockerfile: ", input.DockerFile)
			getPath, err := builtin.FindPath("extracted_file/" + input.AppId)

			if err != nil {
				log4go.Error("Module: StartBuild, MethodName: FindPath, Message: %s ", err.Error())
				log.Println(err)
				return
			}
			if !service.FileExists("extracted_file/" + input.AppId + "/" + getPath + "/Dockerfile") {
				log4go.Info("Module: StartBuild, MethodName: FileExists, Message: Searching for Dockerfile in the extracted_file/" + input.AppId + "/" + getPath + "/Dockerfile" + ". To re-write the dockerfile with the updated dockerfile, which is added on the fly")
				_, err := os.Create("extracted_file/" + input.AppId + "/" + getPath + "/Dockerfile")
				if err != nil {
					log4go.Error("Module: StartBuild, MethodName: Create, Message: %s ", err.Error())
					return
				}
				log4go.Info("Module: StartBuild, MethodName: Create, Message: Docker file doesn't exists in the Extracted file - " + input.AppId + "/" + getPath + "/Dockerfile" + ". Creating a Docker file to add the dockerfile, which is added on the fly")
			}
			data := []byte(input.DockerFile)

			err = ioutil.WriteFile("extracted_file/"+input.AppId+"/"+getPath+"/Dockerfile", data, 0)
			if err != nil {
				log4go.Error("Module: StartBuild, MethodName: WriteFile, Message: %s ", err.Error())
				log.Fatal(err)
				err = service.DeletedSourceFile("extracted_file/" + input.AppId)
				if err != nil {
					log4go.Error("Module: StartBuild, MethodName: DeletedSourceFile, Message: %s ", err.Error())
					helper.RespondwithJSON(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
					return
				}
				log4go.Info("Module: StartBuild, MethodName: DeletedSourceFile, Message: Error occurred while writing the dockerfile. So, Deleting the Source file from the path : extracted_file/" + input.AppId)
			}
			log4go.Info("Module: StartBuild, MethodName: WriteFile, Message: Adding the Docker file content to the dockerfile, which is added on the fly. Dockerfile: " + input.DockerFile)
		}

		filePath, err := service.FindFile("extracted_file/" + input.AppId)

		if filePath == "Docker file doesn't exists" {
			//Delete extracted file and image
			Out.Close()
			err = service.DeletedSourceFile("extracted_file/" + input.AppId)
			if err != nil {
				log4go.Error("Module: StartBuild, MethodName: DeletedSourceFile, Message: %s ", err.Error())
				helper.RespondwithJSON(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
				return
			}

			log4go.Error("Module: StartBuild, MethodName: DeletedSourceFile, Message: Docker file doesn't exists. Deleting the extracted source file - extracted_file/" + input.AppId)
			helper.RespondwithJSON(w, http.StatusBadRequest, map[string]string{"message": filePath})
			return
		}
		if err != nil {
			log.Println(err)
			log4go.Error("Module: StartBuild, MethodName: FindFile, Message: %s ", err.Error())
			Out.Close()
			err = service.DeletedSourceFile("extracted_file/" + input.AppId)
			if err != nil {
				log4go.Error("Module: StartBuild, MethodName: DeletedSourceFile, Message: %s ", err.Error())
				helper.RespondwithJSON(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
				return
			}

			helper.RespondwithJSON(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
			return
		}

		fmt.Println("finished find file" + filePath)

		buildContext, err := remotbuild.NewBuildContext()
		if err != nil {
			log4go.Error("Module: StartBuild, MethodName: NewBuildContext, Message: %s ", err.Error())
			helper.RespondwithJSON(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
			return
		}
		defer buildContext.Close()

		buildContext.AddSourceFolder(filePath, "")
		log4go.Info("Module: StartBuild, MethodName: AddSourceFolder, Message: Adding file to the source folder -" + filePath)
		archive, err := buildContext.Archive()
		if err != nil {
			log4go.Error("Module: StartBuild, MethodName: Archive, Message: %s ", err.Error())
			helper.RespondwithJSON(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
			return
		}
		defer archive.Close()

		ArchiveFile = archive

		source, sourceError := os.Open(archive.Name())
		if sourceError != nil {
			log4go.Error("Module: StartBuild, MethodName: Archive, Message: %s ", sourceError)
			log.Println(sourceError)
		}
		defer source.Close()
		err = os.RemoveAll("extracted_file/" + input.AppId)
		if err != nil {
			log.Println(err)
		}
		log4go.Info("Module: StartBuild, MethodName: RemoveAll, Message: Removed the extracted file from the path - extracted_file/" + input.AppId)

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
		log4go.Error("Module: StartBuild, MethodName: BuildImage, Message: %s ", err.Error())
		_ = os.Remove(input.AppId)
		helper.RespondwithJSON(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
		return
	}
	_ = os.Remove(input.AppId)
	w.Write([]byte(img.Tag))
}
