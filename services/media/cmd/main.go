package main

import (
	"log"

	"youtube-code-backend/pkg/common/config"
	"youtube-code-backend/pkg/common/server"
	"youtube-code-backend/services/media/internal/router"
)

func main() {
	cfg, err := config.FromEnv("media-service", 8004)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	log.Printf("starting %s on :%d", cfg.ServiceName, cfg.Port)
	if err := server.Run(cfg, router.Register); err != nil {
		log.Fatalf("service stopped: %v", err)
	}
}
