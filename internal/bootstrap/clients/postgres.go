package clients

import (
	"fmt"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/austiecodes/dws/lib/resources"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PGConfig struct {
	Host            string `toml:"host"`
	Port            string `toml:"port"`
	User            string `toml:"user"`
	Password        string `toml:"password"`
	DBName          string `toml:"db_name"`
	SSLMode         string `toml:"ssl_mode"`
	MaxOpenConns    int    `toml:"max_open_conns"`
	MaxIdleConns    int    `toml:"max_idle_conns"`
	ConnMaxLifetime int    `toml:"conn_max_lifetime"`
}

type PostgresClient struct {
	config PGConfig
}

func NewPostgresClient() *PostgresClient {
	return &PostgresClient{}
}

func (p *PostgresClient) LoadConfig() error {
	var config struct {
		PG PGConfig `toml:"pg"`
	}
	if _, err := toml.DecodeFile("conf/database.toml", &config); err != nil {
		return fmt.Errorf("error loading database config: %w", err)
	}
	p.config = config.PG
	return nil
}

func (p *PostgresClient) Init() error {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		p.config.Host,
		p.config.Port,
		p.config.User,
		p.config.Password,
		p.config.DBName,
		p.config.SSLMode,
	)

	var err error
	resources.PGClient, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	sqlDB, err := resources.PGClient.DB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	sqlDB.SetMaxOpenConns(p.config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(p.config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(p.config.ConnMaxLifetime) * time.Second)

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	return nil
}
