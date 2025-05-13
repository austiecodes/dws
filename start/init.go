package start

import (
	"fmt"
	"time"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"github.com/austiecodes/dws/lib/managers"
	"github.com/austiecodes/dws/lib/resources"
	"github.com/austiecodes/dws/routes"
	"github.com/docker/docker/client"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var err error

func InitClients(appConfig AppConfig) {
	initLogger(appConfig.Log)
	initDockerClient()
	initPG(appConfig.PG)
	initRabbitMQ(appConfig.MQ)
	// initGPUManager()
}

func InitServer(appConfig AppConfig) {
	config := appConfig.Server
	r := gin.New()
	// load middlewares
	r.Use(ginzap.Ginzap(resources.Logger, time.RFC3339, true))
	r.Use(ginzap.RecoveryWithZap(resources.Logger, true))
	store := cookie.NewStore([]byte(config.SessionKey))
	r.Use(sessions.Sessions(config.SessionName, store))
	// setup routes
	routes.SetupRoutes(r)
	port := fmt.Sprintf(":%d", config.Port)
	if err = r.Run(port); err != nil {
		panic(err)
	}
}

func initLogger(config AppConfigLog) {
	infoLogger := &lumberjack.Logger{
		Filename:   config.InfoLogFilePath,
		MaxSize:    100, // MB
		MaxBackups: 2,
		MaxAge:     28, // days
		Compress:   true,
	}

	warnLogger := &lumberjack.Logger{
		Filename:   config.InfoLogFilePath,
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	}

	errorLogger := &lumberjack.Logger{
		Filename:   config.ErrorLogFilePath,
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	}

	infoSyncer := zapcore.AddSync(infoLogger)
	warnSyncer := zapcore.AddSync(warnLogger)
	errorSyncer := zapcore.AddSync(errorLogger)

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoder := zapcore.NewJSONEncoder(encoderConfig)

	infoCore := zapcore.NewCore(encoder, infoSyncer, zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level == zapcore.InfoLevel
	}))

	warnCore := zapcore.NewCore(encoder, warnSyncer, zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level == zapcore.WarnLevel
	}))

	errorCore := zapcore.NewCore(encoder, errorSyncer, zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level >= zapcore.ErrorLevel
	}))

	core := zapcore.NewTee(infoCore, warnCore, errorCore)
	resources.Logger = zap.New(core, zap.AddCaller())
	defer resources.Logger.Sync()
}

func initDockerClient() {
	resources.DockerClient, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(fmt.Errorf("cannot init docker client: %w", err))
	}
}

func initPG(config AppConfigPG) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host,
		config.User,
		config.Password,
		config.DBName,
		config.SSLMode,
	)

	resources.PGClient, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(fmt.Errorf("failed to connect to PostgreSQL: %w", err))
	}

	// conn pool params
	sqlDB, err := resources.PGClient.DB()
	if err != nil {
		panic(fmt.Errorf("failed to get database connection: %w", err))
	}

	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(config.ConnMaxLifetime) * time.Second)

	// test conn
	if err := sqlDB.Ping(); err != nil {
		panic(fmt.Errorf("failed to ping PostgreSQL: %w", err))
	}
	var dbName string
	resources.PGClient.Raw("SELECT current_database()").Scan(&dbName)
}

func initGPUManager() {
	var errMsg string
	if ret := nvml.Init(); ret != nvml.SUCCESS {
		errMsg = nvml.ErrorString(ret)
		panic(errMsg)
	}

	count, ret := nvml.DeviceGetCount()
	if ret != nvml.SUCCESS {
		errMsg = nvml.ErrorString(ret)
		panic(errMsg)
	}

	resources.GPUManager = &managers.GPUManager{
		Devices: make([]*nvml.Device, count),
	}

	for i := 0; i < count; i++ {
		device, ret := nvml.DeviceGetHandleByIndex(i)
		if ret != nvml.SUCCESS {
			errMsg = nvml.ErrorString(ret)
			panic(errMsg)
		}
		resources.GPUManager.Devices = append(resources.GPUManager.Devices, &device)
	}

}

func initRabbitMQ(config AppConfigMQ) {
	url := fmt.Sprintf("%s://%s:%s@%s:%d",
		config.Protocol,
		config.Username,
		config.Password,
		config.Host,
		config.Port,
	)

	resources.RMQConn, err = amqp091.Dial(url)
	if err != nil {
		panic(fmt.Errorf("failed to init rabbitmq: %w", err))
	}

}
