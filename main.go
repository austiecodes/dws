package main

import (
	"log"

	"github.com/BurntSushi/toml"
	"github.com/austiecodes/dws/start"
)

func main() {
	// parse app config
	appConfigFilePath := "conf/app.toml"
	var appConfig start.AppConfig
	if _, err := toml.DecodeFile(appConfigFilePath, &appConfig); err != nil {
		log.Fatalf("Error parsing config file: %v", err)
	}
	// let's start the app
	start.InitClients(appConfig)
	start.InitServer(appConfig)

}
