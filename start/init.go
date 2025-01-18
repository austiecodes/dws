package start

import (
	"fmt"
	"time"

	"github.com/austiecodes/dws/resources"
	"github.com/docker/docker/client"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var err error

func MustInit() {
	// init docker client
	initDockerClient()
	initPG("conf/pg.toml")
}

func initDockerClient() {
	resources.DockerClient, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(fmt.Errorf("cannot init docker client: %w", err))
	}
}

func initPG(configPath string) {

	config, err := ParsePGConfig(configPath)
	if err != nil {
		panic(fmt.Errorf("failed to parse PostgreSQL config: %w", err))
	}

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

	// conn max life time
	connMaxLifetime, err := time.ParseDuration(config.ConnMaxLifetime)
	if err != nil {
		panic(fmt.Errorf("failed to parse conn max lifetime: %w", err))
	}
	sqlDB.SetConnMaxLifetime(connMaxLifetime)

	// test conn
	if err := sqlDB.Ping(); err != nil {
		panic(fmt.Errorf("failed to ping PostgreSQL: %w", err))
	}
}
