package main

import (
	"log"

	"github.com/BurntSushi/toml"
	"github.com/austiecodes/dws/start"
	// 导入 ginzap
)

func main() {
	// TODO: add config parser here
	// the app init accroding to app.toml
	// bootstrap
	appConfigFilePath := "conf/app.toml"
	var appConfig start.AppConfig
	if _, err := toml.DecodeFile(appConfigFilePath, &appConfig); err != nil {
		log.Fatalf("Error parsing config file: %v", err)
	}

	start.InitClients(appConfig)
	start.InitServer(appConfig)

}
