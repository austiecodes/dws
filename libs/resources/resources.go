package resources

import (
	"github.com/austiecodes/dws/libs/managers"
	"github.com/docker/docker/client"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var DockerClient *client.Client
var PGClient *gorm.DB
var GPUManager *managers.GPUManager
var Logger *zap.Logger
