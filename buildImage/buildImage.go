package buildimage

import (
	"context"
	"log"

	"fmt"
	"io"
	"os"

	 
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
	"start_build/docker"
	"github.com/docker/docker/client"

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

func  BuildImage(ctx context.Context, tar io.Reader, tag string, out io.Writer, buildArgs map[string]interface{}) (*Image, error) {
	opts := types.ImageBuildOptions{
		BuildArgs:   normalizeBuildArgs(buildArgs),
		AuthConfigs: authConfigs(),
		Tags:        []string{tag},
	}

	cli, err := docker.NewDockerClient()
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	resp, err := cli.Client().ImageBuild(ctx, tar, opts)

	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	defer resp.Body.Close()

	termFd, isTerm := term.GetFdInfo(os.Stderr)

	if err := jsonmessage.DisplayJSONMessagesStream(resp.Body, out, termFd, isTerm, nil); err != nil {
		fmt.Println(err.Error())
		return nil, fmt.Errorf("something went wrong with file")
	}

	imgSummary, err := cli.FindImage(ctx, tag)

	
	if err != nil {
		return nil, nil
	}

	image := &Image{
		ID:   imgSummary.ID,
		Tag:  tag,

		Size: imgSummary.Size,
	}

	outs, err := os.Create("debugImage")
	if err != nil {
		return nil, nil
	}
	outs.Close()
	err = cli.PushImage(ctx, image.Tag, outs)
	
	if err != nil {
		return nil,err
	}


	return image, nil
}

func RemoveImage(imageID string)(error){
	cli, err := docker.NewDockerClient()
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	_, err = cli.Client().ImageRemove(context.Background(), imageID,  types.ImageRemoveOptions{Force: true, PruneChildren: true})
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
