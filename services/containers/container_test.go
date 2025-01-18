package services

import (
	"fmt"
	"testing"

	"github.com/austiecodes/dws/models/types"
	"github.com/austiecodes/dws/start"
	"github.com/docker/docker/api/types/container"
	"github.com/gin-gonic/gin"
)

var testContanerID = "5ca6cf69434d"

func TestStopContainers(t *testing.T) {
	start.MustInit()
	ctx := &gin.Context{}
	err := StopContainerService(ctx, testContanerID)
	if err != nil {
		t.Fatal(err)
	}
}

func TestStartContainers(t *testing.T) {
	start.MustInit()
	ctx := &gin.Context{}
	err := StartContainerService(ctx, testContanerID)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRemoveContainers(t *testing.T) {
	start.MustInit()
	containerID := "5ca6cf69434d"
	ctx := &gin.Context{}
	err := RemoveContainerService(ctx, containerID)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateContainers(t *testing.T) {
	start.MustInit()
	config := &types.CreateContainerOptions{
		ContainerConfig: &container.Config{
			Image: "busybox",
		},
		HostConfig: &container.HostConfig{},
	}

	ctx := &gin.Context{}
	resp, err := CreateContainerService(ctx, config)
	fmt.Println(resp)
	if err != nil {
		t.Fatal(err)
	}
}


