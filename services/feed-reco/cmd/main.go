package main

import (
	"log"

	"youtube-code-backend/pkg/common/config"
	"youtube-code-backend/pkg/common/server"
	"youtube-code-backend/services/feed-reco/internal/router"
)

func main() {
	cfg, err := config.FromEnv("feed-reco-service", 8006)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	log.Printf("starting %s on :%d", cfg.ServiceName, cfg.Port)
	if err := server.Run(cfg, router.Register); err != nil {
		log.Fatalf("service stopped: %v", err)
	}
}
