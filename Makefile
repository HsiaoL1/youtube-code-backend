.PHONY: fmt test run-all infra-up infra-down infra-logs infra-ps \
	run-identity run-user-channel run-video run-media run-social-graph run-feed-reco run-comment run-live run-chat run-search run-notification run-moderation run-analytics

fmt:
	gofmt -w $(shell find . -name '*.go' -type f)

test:
	go test ./...

infra-up:
	docker compose up -d postgres mysql mongo redis

infra-down:
	docker compose stop postgres mysql mongo redis

infra-logs:
	docker compose logs -f postgres mysql mongo redis

infra-ps:
	docker compose ps postgres mysql mongo redis

run-identity:
	go run ./services/identity/cmd

run-user-channel:
	go run ./services/user-channel/cmd

run-video:
	go run ./services/video/cmd

run-media:
	go run ./services/media/cmd

run-social-graph:
	go run ./services/social-graph/cmd

run-feed-reco:
	go run ./services/feed-reco/cmd

run-comment:
	go run ./services/comment/cmd

run-live:
	go run ./services/live/cmd

run-chat:
	go run ./services/chat/cmd

run-search:
	go run ./services/search/cmd

run-notification:
	go run ./services/notification/cmd

run-moderation:
	go run ./services/moderation/cmd

run-analytics:
	go run ./services/analytics/cmd

run-all:
	@echo "Run each service in a dedicated terminal using make run-<service>."
