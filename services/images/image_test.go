package services_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/BurntSushi/toml"
	services "github.com/austiecodes/dws/services/images"
	"github.com/austiecodes/dws/start"
	"github.com/gin-gonic/gin"
)

var appConfigFilePath = "../../conf/app.toml"
var appConfig start.AppConfig

func parseConfig() {
	if _, err := toml.DecodeFile(appConfigFilePath, &appConfig); err != nil {
		log.Fatalf("Error parsing config file: %v", err)
	}
}

func TestListImage(t *testing.T) {
	parseConfig()
	start.InitClients(appConfig)

	ctx := &gin.Context{}
	imgs, err := services.ListImages(ctx)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(imgs)
}
