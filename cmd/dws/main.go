package main

import (
	"log"

	"github.com/austiecodes/dws/internal/bootstrap"
)

func main() {
	if err := bootstrap.Bootstrap(); err != nil {
		log.Fatalf("Failed to bootstrap application: %v", err)
	}
}
