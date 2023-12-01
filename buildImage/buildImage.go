package buildimage

import (
	"bufio"
	"context"
	"log"
	"strconv"

	"fmt"
	"io"
	"os"

	"start_build/docker"

	"github.com/alecthomas/log4go"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
)

type DockerClient struct {
	docker       *client.Client
	registryAuth string
}

type Image struct {
	ID   string
	Tag  string
	Size int64
}

func (c *DockerClient) Client() *client.Client {
	return c.docker
}

type DeployOperation struct {
	DockerClient    *DockerClient
	DockerAvailable bool
	Out             io.Writer
	AppName         string
	ImageTag        string
	RemoteOnly      bool
	LocalOnly       bool
}

func authConfigs() map[string]types.AuthConfig {
	authConfigs := map[string]types.AuthConfig{}

	dockerhubUsername := os.Getenv("DOCKER_HUB_USERNAME")
	dockerhubPassword := os.Getenv("DOCKER_HUB_PASSWORD")

	if dockerhubUsername != "" && dockerhubPassword != "" {
		cfg := types.AuthConfig{
			Username:      dockerhubUsername,
			Password:      dockerhubPassword,
			ServerAddress: "index.docker.io",
		}
		authConfigs["https://index.docker.io/v1/"] = cfg
	}

	return authConfigs
}

func BuildImage(ctx context.Context, tar io.Reader, tag string, out io.Writer, buildArgs map[string]interface{}, dockerFileName string) (*Image, []string, error) {

	fmt.Println("------------------", tag)
	fmt.Println("------------------", tag)
	fmt.Println("------------------", tag)
	fmt.Println("------------------", tag)
	fmt.Println("------------------", tag)

	fmt.Println("Started remote build --------------------------------------------------------")
	opts := types.ImageBuildOptions{
		BuildArgs:   normalizeBuildArgs(buildArgs),
		AuthConfigs: authConfigs(),
		Tags:        []string{tag},
		Dockerfile:  dockerFileName,
	}
	fmt.Println("Docker File Name--------------------------------------------------------", dockerFileName)

	cli, err := docker.NewDockerClient()
	if err != nil {
		log4go.Error("Module: StartBuild, MethodName: NewDockerClient, Message: %s ", err.Error())
		fmt.Println(err.Error())
		return nil, []string{}, err
	}
	fmt.Println("NewDockerClient --------------------------------------------------------")

	var buildLogs []string

	buildLogs = []string{"Checking the uploaded file format...", "Extracting the file..."}

	resp, err := cli.Client().ImageBuild(ctx, tar, opts)

	if err != nil {
		fmt.Println("Unable to retrieve build logs.")
		log4go.Error("Module: StartBuild, MethodName: ImageBuild, Message: %s ", err.Error())
		fmt.Println(err.Error())
		return nil, []string{}, err
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "{\"stream\":\"\\n\"}" {
			fmt.Println(line)
			buildLogs = append(buildLogs, line)
		}
	}
	fmt.Println("response --------------------------------------------------------", resp.Body)

	log4go.Info("Module: StartBuild, MethodName: ImageBuild, Message: Building the file as docker image")
	defer resp.Body.Close()

	termFd, isTerm := term.GetFdInfo(os.Stderr)

	if err := jsonmessage.DisplayJSONMessagesStream(resp.Body, out, termFd, isTerm, nil); err != nil {
		fmt.Println(err.Error())
		log4go.Error("Module: StartBuild, MethodName: DisplayJSONMessagesStream, Message: %s ", err.Error())
		return nil, []string{}, fmt.Errorf("something went wrong with file")
	}
	fmt.Println("DisplayJSONMessagesStream --------------------------------------------------------")

	imgSummary, err := cli.FindImage(ctx, tag)

	if err != nil {
		log4go.Error("Module: StartBuild, MethodName: FindImage, Message: %s ", err.Error())
		return nil, []string{}, err
	}
	log4go.Info("Module: StartBuild, MethodName: FindImage, Message: Finding Image using the tag - " + tag + " . The size of the Image - " + strconv.Itoa(int(imgSummary.Size)))
	fmt.Println("FindImage --------------------------------------------------------", imgSummary.ID, "---SIZE----", imgSummary.Size)

	image := &Image{
		ID:  imgSummary.ID,
		Tag: tag,

		Size: imgSummary.Size,
	}

	outs, err := os.Create("debugImage")
	if err != nil {
		return nil, []string{}, err
	}
	outs.Close()
	err = cli.PushImage(ctx, image.Tag, outs)

	if err != nil {
		log4go.Error("Module: StartBuild, MethodName: PushImage, Message: %s ", err.Error())
		return nil, []string{}, err
	}

	log4go.Info("Module: StartBuild, MethodName: PushImage, Message: Docker Image - " + image.Tag)
	fmt.Println("PushImage --------------------------------------------------------", image)

	return image, buildLogs, nil
}

func RemoveImage(imageID string) error {
	cli, err := docker.NewDockerClient()
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	_, err = cli.Client().ImageRemove(context.Background(), imageID, types.ImageRemoveOptions{Force: true, PruneChildren: true})
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func normalizeBuildArgs(buildArgs map[string]interface{}) map[string]*string {
	var out = map[string]*string{}
	for k, v := range buildArgs {
		out[k] = v.(*string)
	}
	return out
}
