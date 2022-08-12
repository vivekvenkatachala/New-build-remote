package docker

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"os"
	"regexp"
	"strings"

	"start_build/terminal"

	"github.com/alecthomas/log4go"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/moby/term"

	dockerparser "github.com/novln/docker-parser"
)

type DockerClient struct {
	docker       *client.Client
	registryAuth string
}

func (c *DockerClient) Client() *client.Client {
	return c.docker
}

func NewDockerClient() (*DockerClient, error) {
	cli, err := client.NewClientWithOpts(client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	if err := client.FromEnv(cli); err != nil {
		return nil, err
	}

	//accessToken := flyctl.GetAPIToken()
	authConfig := RegistryAuth("")
	encodedJSON, err := json.Marshal(authConfig)
	if err != nil {
		return nil, err
	}
	authStr := base64.URLEncoding.EncodeToString(encodedJSON)
	c := &DockerClient{
		docker:       cli,
		registryAuth: authStr,
	}

	return c, nil
}

func RegistryAuth(token string) types.AuthConfig {
	return types.AuthConfig{
		Username:      "nife123",
		Password:      "Nife@2020",
		ServerAddress: "hub.docker.com",
	}
}

var imageIDPattern = regexp.MustCompile("[a-f0-9]")

func (c *DockerClient) FindImage(ctx context.Context, imageName string) (*types.ImageSummary, error) {
	ref, err := dockerparser.Parse(imageName)
	if err != nil {
		return nil, err
	}

	isID := imageIDPattern.MatchString(imageName)

	images, err := c.docker.ImageList(ctx, types.ImageListOptions{})
	if err != nil {
		return nil, err
	}

	if isID {
		for _, img := range images {
			if len(img.ID) < len(imageName)+7 {
				continue
			}
			if img.ID[7:7+len(imageName)] == imageName {
				terminal.Debug("Found image by id", imageName)
				return &img, nil
			}
		}
	}

	searchTerms := []string{
		imageName,
		imageName + ":" + ref.Tag(),
		ref.Name(),
		ref.ShortName(),
		ref.Remote(),
		ref.Repository(),
	}

	terminal.Debug("Search terms:", searchTerms)

	for _, img := range images {
		for _, tag := range img.RepoTags {
			// skip <none>:<none>
			if strings.HasPrefix(tag, "<none>") {
				continue
			}

			for _, term := range searchTerms {
				if tag == term {
					return &img, nil
				}
			}
		}
	}

	return nil, nil
}

func (c *DockerClient) PushImage(ctx context.Context, imageName string, out io.Writer) error {
	resp, err := c.docker.ImagePush(ctx, imageName, types.ImagePushOptions{RegistryAuth: c.registryAuth})
	if err != nil {
		log4go.Error("Module: StartBuild, MethodName: PushImage, Message: %s ", err.Error())
		return err
	}
	defer resp.Close()

	termFd, isTerm := term.GetFdInfo(os.Stderr)
	return jsonmessage.DisplayJSONMessagesStream(resp, out, termFd, isTerm, nil)
}
