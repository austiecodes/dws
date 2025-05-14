package clients

import (
	"fmt"

	"github.com/austiecodes/dws/lib/resources"
	"github.com/docker/docker/client"
)

type DockerClient struct{}

func NewDockerClient() *DockerClient {
	return &DockerClient{}
}

// LoadConfig is a no-op for Docker client as it uses environment variables
func (d *DockerClient) LoadConfig() error {
	return nil
}

func (d *DockerClient) Init() error {
	var err error
	resources.DockerClient, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("cannot init docker client: %w", err)
	}
	return nil
}
