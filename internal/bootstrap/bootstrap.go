package bootstrap

import (
	"fmt"
	"log"

	"github.com/austiecodes/dws/internal/bootstrap/clients"
	"github.com/austiecodes/dws/internal/server"
)

// Bootstrap initializes all application components
func Bootstrap() error {
	// Create all initializers
	initializers := []Initializer{
		clients.NewLoggerClient(),   // Initialize logger first
		clients.NewPostgresClient(), // Then database
		clients.NewRabbitMQClient(), // Then message queue
		clients.NewRedisClient(),    // Then Redis
		clients.NewDockerClient(),   // Then Docker
		clients.NewGPUClient(),      // Then GPU
	}

	// Initialize all clients
	if err := InitializeAll(initializers...); err != nil {
		return fmt.Errorf("failed to initialize clients: %w", err)
	}

	// Initialize and start server
	srv := server.NewServer()
	if err := srv.LoadConfig(); err != nil {
		return fmt.Errorf("failed to load server config: %w", err)
	}
	if err := srv.Init(); err != nil {
		return fmt.Errorf("failed to initialize server: %w", err)
	}
	log.Println("All components initialized successfully")
	return srv.Start()
}
