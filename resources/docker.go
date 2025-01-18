package resources

import (
	"github.com/docker/docker/client"
	"gorm.io/gorm"
)

var DockerClient *client.Client
var PGClient *gorm.DB
