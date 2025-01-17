package start

import (
	"log"

	"github.com/austiecodes/dws/resources"
	"github.com/docker/docker/client"
)

func MustInit() {
	// init docker client
	initDockerClient()
}

var err error

func initDockerClient() {
	resources.DockerClient, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("cannot init docker client: %v", err)
	}
}
