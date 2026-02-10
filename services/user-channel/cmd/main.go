package main

import (
	"log"

	"youtube-code-backend/pkg/common/config"
	"youtube-code-backend/pkg/common/server"
	"youtube-code-backend/services/user-channel/internal/router"
)

func main() {
	cfg, err := config.FromEnv("user-channel-service", 8002)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	log.Printf("starting %s on :%d", cfg.ServiceName, cfg.Port)
	if err := server.Run(cfg, router.Register); err != nil {
		log.Fatalf("service stopped: %v", err)
	}
}
