# youtube-code-backend

Go + Fiber microservice skeleton for a YouTube-like backend.

## Services
- `identity` on `:8001`
- `user-channel` on `:8002`
- `video` on `:8003`
- `media` on `:8004`
- `social-graph` on `:8005`
- `feed-reco` on `:8006`
- `comment` on `:8007`
- `live` on `:8008`
- `chat` on `:8009`
- `search` on `:8010`
- `notification` on `:8011`
- `moderation` on `:8012`
- `analytics` on `:8013`

## Service Responsibilities

### 1) `identity-service` (`services/identity`)
- Login/register (email/phone/username).
- Access + refresh token issuance and refresh flow.
- Logout, token invalidation, and force logout.
- Role and permission model (`user`/`creator`/`admin`).
- RBAC middleware (route-level and resource-level).
- Account security basics: password reset/change hooks, verification code hooks.

### 2) `user-channel-service` (`services/user-channel`)
- User profile read/update: avatar, nickname, bio, region, links.
- Account state exposure: normal/disabled/banned with reason/expiry.
- Channel profile management: banner, intro, basic stats.
- Channel owner permissions for editing.
- Channel content aggregation entry points (videos/shorts/live replay/playlists).

### 3) `video-service` (`services/video`)
- Video metadata CRUD for long/short/replay (shared model + `type`).
- Video lifecycle state machine: `draft/uploading/processing/ready/rejected/removed`.
- Visibility rules: `public/unlisted/private/scheduled`.
- Publish now / schedule publish.
- Interaction domain for video: like/unlike, favorite/unfavorite.
- Playlist domain: create/update/reorder/add/remove videos.
- Playback info API integration surface (resolve playable metadata).

### 4) `media-service` (`services/media`)
- Upload session creation and upload credential issuing.
- Multipart/chunked/resumable upload orchestration.
- Upload-complete callback handling.
- Async media pipeline trigger: transcode/snapshot/metadata probe/subtitle bind.
- HLS output management (minimum viable stream packaging).
- Signed media URL generation and file access auth.
- Integration hooks to object storage + CDN.

### 5) `social-graph-service` (`services/social-graph`)
- Subscribe/follow and unfollow channel APIs.
- Following list and follower list queries.
- Follow graph lookups for other services (feed/notification).
- Event emission when relationship changes (for feed/notifications).

### 6) `feed-reco-service` (`services/feed-reco`)
- Home feed aggregation (`hot + subscriptions + recent`).
- Subscription feed (time-order baseline).
- Trending feed (24h/7d weighted score baseline).
- Category feed and shorts feed endpoints.
- Pagination and preload-friendly feed contracts.
- Recommendation hooks (related videos, creator-more).

### 7) `comment-service` (`services/comment`)
- Comment list by time/hot sorting.
- Create comment/reply (two-level thread model).
- Delete (owner/admin), pin (creator/admin), like comment.
- Basic anti-spam/rate-limit/sensitive-word checks.
- Report-entry hook for moderation integration.

### 8) `live-service` (`services/live`)
- Live room create/update (`title/category/cover/scheduledAt`).
- Stream key and publish URL generation.
- Live state transitions: `scheduled/live/ended`.
- Live playback source APIs (HLS).
- Stream auth hooks (publish/play auth policy integration).
- Replay generation trigger after stream ends (to VOD flow).
- Near-real-time counters: online/heat (can be approximate).

### 9) `chat-service` (`services/chat`)
- WebSocket chat room connection lifecycle.
- Send/receive chat messages and pull recent N history.
- Room-level moderation features: mute, manager roles, keyword filter hooks.
- Chat event stream hooks (join/follow/gift placeholder events).
- Message rate-limiting and abuse-protection hooks.

### 10) `search-service` (`services/search`)
- Video/channel/live search endpoints.
- Basic ranking modes: relevant/latest/most-viewed.
- Search index update consumers (from domain events).
- Initial implementation via DB full-text search; extensible to ES/Meilisearch.

### 11) `notification-service` (`services/notification`)
- In-app notification center (list, unread/read).
- Notification types: publish/live/comment reply/@mention/moderation result.
- Event-driven notification production and fanout.
- Optional sender adapters: email/SMS/push (post-MVP).

### 12) `moderation-service` (`services/moderation`)
- Content moderation status management (video/short/live/profile/comment/chat).
- Manual review workflow: queue, decision, rejection reason, re-submit.
- Report/ticket workflow: `open/in_progress/resolved/rejected`.
- Enforcement actions: remove/takedown/ban/warn/rank-limit hooks.
- Admin audit logs for all moderation actions.

### 13) `analytics-service` (`services/analytics`)
- Event ingestion: play start/end, watch duration, completion, interactions.
- Counters: PV/UV, likes/comments/favorites/shares.
- Aggregation: daily creator metrics and dashboard-facing summaries.
- Live metrics: peak concurrent viewers, total viewers, chat volume.
- Async metric compute jobs and backfill/recompute hooks.

## Cross-service platform expectations
- Consistent request ID, structured logging, and unified error code style.
- Rate limiting and anti-abuse policy hooks per service boundary.
- Idempotency support for create/upload style endpoints.
- Async task/event integration for transcode, notification, index, analytics.

Each service exposes:
- `GET /healthz`
- `GET /readyz`
- `GET /api/v1/<service>/ping`
- `GET /api/v1/<service>/todo`

## Quick start
1. Install deps:
   `go mod tidy`
2. Start local infra (PostgreSQL/MySQL/MongoDB/Redis):
   `make infra-up`
3. Check infra status:
   `make infra-ps`
4. Run one service:
   `make run-video`
5. Format and test:
   `make fmt && make test`

## Local infrastructure defaults
- PostgreSQL
  - Host: `127.0.0.1`
  - Port: `5432`
  - DB/User/Password: `youtube / youtube / youtube`
  - DSN: `host=127.0.0.1 user=youtube password=youtube dbname=youtube port=5432 sslmode=disable`
- MySQL
  - Host: `127.0.0.1`
  - Port: `3306`
  - DB/User/Password: `youtube / dev / dev`
  - Root password: `root`
  - DSN: `dev:dev@tcp(127.0.0.1:3306)/youtube?charset=utf8mb4&parseTime=True&loc=Local`
- MongoDB
  - Host: `127.0.0.1`
  - Port: `27017`
  - Root user/password: `dev / dev`
  - URI: `mongodb://dev:dev@127.0.0.1:27017/youtube?authSource=admin`
- Redis
  - Host: `127.0.0.1`
  - Port: `6379`
  - Password: empty
  - URL: `redis://127.0.0.1:6379/0`
