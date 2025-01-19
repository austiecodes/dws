package services_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/austiecodes/dws/models/types"
	services "github.com/austiecodes/dws/services/containers"
	"github.com/austiecodes/dws/start"
	"github.com/docker/docker/api/types/container"
	"github.com/gin-gonic/gin"
)

var testContanerID = "5ca6cf69434d"

var appConfigFilePath = "../../conf/app.toml"
var appConfig start.AppConfig

func parseConfig() {
	if _, err := toml.DecodeFile(appConfigFilePath, &appConfig); err != nil {
		log.Fatalf("Error parsing config file: %v", err)
	}
}

func TestStopContainers(t *testing.T) {
	parseConfig()
	start.InitClients(appConfig)
	ctx := &gin.Context{}
	err := services.StopContainerService(ctx, testContanerID)
	if err != nil {
		t.Fatal(err)
	}
}

func TestStartContainers(t *testing.T) {
	parseConfig()
	start.InitClients(appConfig)
	ctx := &gin.Context{}
	err := services.StartContainerService(ctx, testContanerID)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRemoveContainers(t *testing.T) {
	parseConfig()
	start.InitClients(appConfig)
	containerID := "5ca6cf69434d"
	ctx := &gin.Context{}
	err := services.RemoveContainerService(ctx, containerID)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateContainers(t *testing.T) {
	parseConfig()
	start.InitClients(appConfig)
	config := &types.CreateContainerOptions{
		ContainerConfig: &container.Config{
			Image: "busybox",
		},
		HostConfig: &container.HostConfig{},
	}

	ctx := &gin.Context{}
	resp, err := services.CreateContainerService(ctx, config)
	fmt.Println(resp)
	if err != nil {
		t.Fatal(err)
	}
}
