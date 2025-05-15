package clients

import (
	"fmt"

	"github.com/austiecodes/dws/lib/resources"
	"github.com/rabbitmq/amqp091-go"
)

type MQConfig struct {
	Protocol string `toml:"protocol"`
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	Username string `toml:"username"`
	Password string `toml:"password"`
}

type RabbitMQClient struct {
	config MQConfig
}

func NewRabbitMQClient() *RabbitMQClient {
	return &RabbitMQClient{}
}

func (r *RabbitMQClient) LoadConfig() error {
	var config struct {
		MQ MQConfig `toml:"mq"`
	}
	if err := LoadConfig("mq.toml", &config); err != nil {
		return fmt.Errorf("error loading MQ config: %w", err)
	}
	r.config = config.MQ
	return nil
}

func (r *RabbitMQClient) Init() error {
	url := fmt.Sprintf("%s://%s:%s@%s:%d",
		r.config.Protocol,
		r.config.Username,
		r.config.Password,
		r.config.Host,
		r.config.Port,
	)

	var err error
	resources.RMQConn, err = amqp091.Dial(url)
	if err != nil {
		return fmt.Errorf("failed to init rabbitmq: %w", err)
	}
	return nil
}
