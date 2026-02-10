# Service Map

## Core services (10)
- `identity-service`: auth, token, RBAC
- `user-channel-service`: user profile, channel
- `video-service`: long/short/replay metadata and publish lifecycle
- `media-service`: upload session, callback, transcode pipeline hooks
- `social-graph-service`: subscribe/follow graph
- `feed-reco-service`: home/subscription/trending feed
- `comment-service`: comments and replies
- `live-service`: live room and stream state
- `chat-service`: websocket chat
- `search-service`: search query and indexing APIs

## Platform services (3)
- `notification-service`: in-app notification production/consumption
- `moderation-service`: moderation and reports workflow
- `analytics-service`: event ingestion and aggregation

## Shared package
- `/Users/hsiaol1/code/youtube-code-backend/pkg/common/config/config.go`
- `/Users/hsiaol1/code/youtube-code-backend/pkg/common/server/server.go`

## Route shape
Each service currently exposes:
- `GET /healthz`
- `GET /readyz`
- `GET /api/v1/<service>/ping`
- `GET /api/v1/<service>/todo`

Add real routes in each service router file under:
- `/Users/hsiaol1/code/youtube-code-backend/services/<service>/internal/router/router.go`

Add business logic in:
- `/Users/hsiaol1/code/youtube-code-backend/services/<service>/internal/handler/handler.go`
