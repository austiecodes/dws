package clients

import (
	"context"
	"fmt"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/austiecodes/dws/lib/resources"
	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Host            string `toml:"host"`
	Port            string `toml:"port"`
	Password        string `toml:"password"`
	DB              int    `toml:"db"`
	PoolSize        int    `toml:"pool_size"`
	DialTimeout     int    `toml:"dial_timeout"`
	ReadTimeout     int    `toml:"read_timeout"`
	WriteTimeout    int    `toml:"write_timeout"`
	ConnMaxLifetime int    `toml:"conn_max_lifetime"`
}

type RedisClient struct {
	config RedisConfig
}

func NewRedisClient() *RedisClient {
	return &RedisClient{}
}

func (r *RedisClient) LoadConfig() error {
	var config struct {
		Redis RedisConfig `toml:"redis"`
	}
	if _, err := toml.DecodeFile("conf/redis.toml", &config); err != nil {
		return fmt.Errorf("error loading redis config: %w", err)
	}
	r.config = config.Redis
	return nil
}

func (r *RedisClient) Init() error {
	addr := fmt.Sprintf("%s:%s", r.config.Host, r.config.Port)

	resources.RedisClient = redis.NewClient(&redis.Options{
		Addr:            addr,
		Password:        r.config.Password,
		DB:              r.config.DB,
		PoolSize:        r.config.PoolSize,
		DialTimeout:     time.Duration(r.config.DialTimeout) * time.Second,
		ReadTimeout:     time.Duration(r.config.ReadTimeout) * time.Second,
		WriteTimeout:    time.Duration(r.config.WriteTimeout) * time.Second,
		ConnMaxLifetime: time.Duration(r.config.ConnMaxLifetime) * time.Second,
	})

	// Test the connection
	ctx := context.Background()
	if err := resources.RedisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return nil
}
