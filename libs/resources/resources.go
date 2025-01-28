package resources

import (
	"github.com/austiecodes/dws/libs/managers"
	"github.com/docker/docker/client"
	"github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var DockerClient *client.Client
var PGClient *gorm.DB
var GPUManager *managers.GPUManager
var Logger *zap.Logger
var RedisClient *redis.Client
var RMQConn *amqp091.Connection
