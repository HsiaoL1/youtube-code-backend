package main

import (
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"youtube-code-backend/pkg/common/config"
	"youtube-code-backend/pkg/common/database"
)

func main() {
	cfg, err := config.FromEnv("gateway-seed", 8000)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := database.Connect(cfg.DatabaseDSN)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}

	log.Println("Running migrations...")
	runMigrations(db)

	log.Println("Adding extra columns...")
	addExtraColumns(db)

	log.Println("Seeding data...")
	seedData(db)

	log.Println("Seed complete!")
}

func runMigrations(db *gorm.DB) {
	// Create all tables if they don't exist
	tables := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id BIGSERIAL PRIMARY KEY,
			username VARCHAR(50) NOT NULL UNIQUE,
			email VARCHAR(255) NOT NULL UNIQUE,
			phone VARCHAR(20) DEFAULT '',
			password_hash VARCHAR(255) NOT NULL,
			role VARCHAR(20) NOT NULL DEFAULT 'user',
			status VARCHAR(20) NOT NULL DEFAULT 'active',
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW(),
			deleted_at TIMESTAMPTZ
		)`,
		`CREATE TABLE IF NOT EXISTS user_profiles (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL UNIQUE,
			nickname VARCHAR(100) DEFAULT '',
			avatar VARCHAR(500) DEFAULT '',
			bio VARCHAR(1000) DEFAULT '',
			region VARCHAR(50) DEFAULT '',
			gender VARCHAR(20) DEFAULT '',
			links TEXT DEFAULT '[]',
			account_status VARCHAR(20) NOT NULL DEFAULT 'active',
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW(),
			deleted_at TIMESTAMPTZ
		)`,
		`CREATE TABLE IF NOT EXISTS channels (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL,
			handle VARCHAR(100) NOT NULL UNIQUE,
			name VARCHAR(200) NOT NULL,
			description VARCHAR(2000) DEFAULT '',
			banner VARCHAR(500) DEFAULT '',
			subscriber_count BIGINT NOT NULL DEFAULT 0,
			video_count BIGINT NOT NULL DEFAULT 0,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW(),
			deleted_at TIMESTAMPTZ
		)`,
		`CREATE TABLE IF NOT EXISTS videos (
			id BIGSERIAL PRIMARY KEY,
			channel_id BIGINT NOT NULL,
			type VARCHAR(20) NOT NULL DEFAULT 'video',
			title VARCHAR(255) NOT NULL,
			description TEXT DEFAULT '',
			status VARCHAR(20) NOT NULL DEFAULT 'draft',
			visibility VARCHAR(20) NOT NULL DEFAULT 'private',
			duration BIGINT DEFAULT 0,
			view_count BIGINT DEFAULT 0,
			like_count BIGINT DEFAULT 0,
			dislike_count BIGINT DEFAULT 0,
			comment_count BIGINT DEFAULT 0,
			favorite_count BIGINT DEFAULT 0,
			thumbnail_url VARCHAR(500) DEFAULT '',
			scheduled_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW(),
			deleted_at TIMESTAMPTZ
		)`,
		`CREATE TABLE IF NOT EXISTS video_likes (
			id BIGSERIAL PRIMARY KEY,
			video_id BIGINT NOT NULL,
			user_id BIGINT NOT NULL,
			is_like BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW(),
			deleted_at TIMESTAMPTZ,
			UNIQUE(video_id, user_id)
		)`,
		`CREATE TABLE IF NOT EXISTS video_favorites (
			id BIGSERIAL PRIMARY KEY,
			video_id BIGINT NOT NULL,
			user_id BIGINT NOT NULL,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW(),
			deleted_at TIMESTAMPTZ,
			UNIQUE(video_id, user_id)
		)`,
		`CREATE TABLE IF NOT EXISTS comments (
			id BIGSERIAL PRIMARY KEY,
			video_id BIGINT NOT NULL,
			user_id BIGINT NOT NULL,
			parent_id BIGINT,
			content TEXT NOT NULL,
			like_count BIGINT DEFAULT 0,
			reply_count BIGINT DEFAULT 0,
			is_pinned BOOLEAN DEFAULT false,
			is_hearted BOOLEAN DEFAULT false,
			status VARCHAR(20) NOT NULL DEFAULT 'active',
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW(),
			deleted_at TIMESTAMPTZ
		)`,
		`CREATE TABLE IF NOT EXISTS comment_likes (
			id BIGSERIAL PRIMARY KEY,
			comment_id BIGINT NOT NULL,
			user_id BIGINT NOT NULL,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW(),
			deleted_at TIMESTAMPTZ,
			UNIQUE(comment_id, user_id)
		)`,
		`CREATE TABLE IF NOT EXISTS live_rooms (
			id BIGSERIAL PRIMARY KEY,
			channel_id BIGINT NOT NULL,
			title VARCHAR(255) NOT NULL,
			description TEXT DEFAULT '',
			status VARCHAR(20) NOT NULL DEFAULT 'idle',
			stream_key VARCHAR(255) NOT NULL DEFAULT '',
			publish_url VARCHAR(500) DEFAULT '',
			playback_url VARCHAR(500) DEFAULT '',
			viewer_count BIGINT DEFAULT 0,
			peak_viewer_count BIGINT DEFAULT 0,
			category VARCHAR(100) DEFAULT '',
			thumbnail_url VARCHAR(500) DEFAULT '',
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW(),
			deleted_at TIMESTAMPTZ
		)`,
		`CREATE TABLE IF NOT EXISTS chat_messages (
			id BIGSERIAL PRIMARY KEY,
			room_id BIGINT NOT NULL,
			user_id BIGINT NOT NULL,
			type VARCHAR(20) NOT NULL DEFAULT 'text',
			content TEXT NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'active',
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW(),
			deleted_at TIMESTAMPTZ
		)`,
		`CREATE TABLE IF NOT EXISTS subscriptions (
			id BIGSERIAL PRIMARY KEY,
			subscriber_id BIGINT NOT NULL,
			channel_id BIGINT NOT NULL,
			notify_preference VARCHAR(50) NOT NULL DEFAULT 'all',
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW(),
			deleted_at TIMESTAMPTZ,
			UNIQUE(subscriber_id, channel_id)
		)`,
		`CREATE TABLE IF NOT EXISTS playlists (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL,
			title VARCHAR(255) NOT NULL,
			description TEXT DEFAULT '',
			visibility VARCHAR(20) NOT NULL DEFAULT 'private',
			video_count BIGINT DEFAULT 0,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW(),
			deleted_at TIMESTAMPTZ
		)`,
		`CREATE TABLE IF NOT EXISTS playlist_items (
			id BIGSERIAL PRIMARY KEY,
			playlist_id BIGINT NOT NULL,
			video_id BIGINT NOT NULL,
			position INT NOT NULL DEFAULT 0,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW(),
			deleted_at TIMESTAMPTZ
		)`,
		`CREATE TABLE IF NOT EXISTS reports (
			id BIGSERIAL PRIMARY KEY,
			reporter_id BIGINT DEFAULT 0,
			content_type VARCHAR(50) NOT NULL,
			content_id BIGINT NOT NULL,
			reason VARCHAR(255) NOT NULL,
			description TEXT DEFAULT '',
			status VARCHAR(20) DEFAULT 'open',
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW(),
			deleted_at TIMESTAMPTZ
		)`,
	}

	for _, t := range tables {
		if err := db.Exec(t).Error; err != nil {
			log.Printf("migration warning: %v", err)
		}
	}
}

func addExtraColumns(db *gorm.DB) {
	// Add tags, category, hls_url columns to videos if they don't exist
	extras := []string{
		`ALTER TABLE videos ADD COLUMN IF NOT EXISTS tags TEXT DEFAULT ''`,
		`ALTER TABLE videos ADD COLUMN IF NOT EXISTS category VARCHAR(100) DEFAULT ''`,
		`ALTER TABLE videos ADD COLUMN IF NOT EXISTS hls_url VARCHAR(500) DEFAULT ''`,
	}
	for _, q := range extras {
		if err := db.Exec(q).Error; err != nil {
			log.Printf("add column warning: %v", err)
		}
	}
}

func seedData(db *gorm.DB) {
	now := time.Now()
	day := 24 * time.Hour
	defaultHLS := "https://test-streams.mux.dev/x36xhzz/x36xhzz.m3u8"

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	hash := string(passwordHash)

	// Clear existing data
	for _, table := range []string{
		"playlist_items", "playlists", "comment_likes", "comments",
		"video_likes", "video_favorites", "chat_messages", "live_rooms",
		"subscriptions", "reports", "videos", "channels", "user_profiles", "users",
	} {
		db.Exec(fmt.Sprintf("DELETE FROM %s", table))
		// Reset sequence
		db.Exec(fmt.Sprintf("ALTER SEQUENCE IF EXISTS %s_id_seq RESTART WITH 1", table))
	}

	// === Users ===
	type userSeed struct {
		id       uint64
		username string
		email    string
		role     string
		nickname string
		avatar   string
		handle   string
		subCount int64
	}
	users := []userSeed{
		{1, "user", "user@example.com", "user", "Alice User", "https://i.pravatar.cc/120?img=12", "alice", 1200},
		{2, "creator", "creator@example.com", "creator", "Chris Creator", "https://i.pravatar.cc/120?img=24", "creator", 98000},
		{3, "admin", "admin@example.com", "admin", "Ada Admin", "https://i.pravatar.cc/120?img=36", "admin", 3200},
		{4, "nora", "nora@example.com", "creator", "Nora Stream", "https://i.pravatar.cc/120?img=48", "nora", 45000},
		{5, "maxgaming", "max@example.com", "creator", "Max Gaming", "https://i.pravatar.cc/120?img=5", "maxgaming", 320000},
		{6, "melodyvibes", "melody@example.com", "creator", "Melody Vibes", "https://i.pravatar.cc/120?img=15", "melodyvibes", 1500000},
		{7, "newsdaily", "news@example.com", "creator", "News Daily", "https://i.pravatar.cc/120?img=33", "newsdaily", 890000},
		{8, "sportzone", "sport@example.com", "creator", "Sport Zone", "https://i.pravatar.cc/120?img=57", "sportzone", 670000},
	}

	for _, u := range users {
		db.Exec(`INSERT INTO users (id, username, email, password_hash, role, status, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, 'active', ?, ?)
			ON CONFLICT (id) DO NOTHING`, u.id, u.username, u.email, hash, u.role, now, now)
		db.Exec(`INSERT INTO user_profiles (user_id, nickname, avatar, account_status, created_at, updated_at)
			VALUES (?, ?, ?, 'active', ?, ?)
			ON CONFLICT (user_id) DO NOTHING`, u.id, u.nickname, u.avatar, now, now)
		db.Exec(`INSERT INTO channels (id, user_id, handle, name, subscriber_count, video_count, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, 0, ?, ?)
			ON CONFLICT (id) DO NOTHING`, u.id, u.id, u.handle, u.nickname, u.subCount, now, now)
	}
	// Reset user sequence past our seeded IDs
	db.Exec("SELECT setval('users_id_seq', 100, true)")
	db.Exec("SELECT setval('user_profiles_id_seq', 100, true)")
	db.Exec("SELECT setval('channels_id_seq', 100, true)")

	cover := func(seed string) string {
		return fmt.Sprintf("https://picsum.photos/seed/%s/640/360", seed)
	}

	// === Videos (long) ===
	type videoSeed struct {
		id          uint64
		channelID   uint64
		vType       string
		title       string
		description string
		cover       string
		duration    int64
		views       int64
		likes       int64
		tags        string
		category    string
		daysAgo     int
		hlsURL      string
	}
	longVideos := []videoSeed{
		{1, 2, "video", "React 18 in Production: Battle-tested Patterns", "Deep dive into data, route and rendering patterns. We cover Suspense, transitions, server components, and how to structure large React applications for performance and maintainability.", cover("react18"), 1330, 560000, 22000, `["react","frontend"]`, "Tech", 3, defaultHLS},
		{2, 4, "video", "How We Designed a Modern Video UX", "Case study on feed, playback, and creator flow. Learn about the design decisions behind building a video platform from scratch.", cover("ux"), 920, 210000, 8000, `["ux","product"]`, "Design", 6, defaultHLS},
		{3, 5, "video", "Top 10 Indie Games You Missed in 2024", "Hidden gems from the indie scene that deserve your attention. From puzzle platformers to narrative adventures.", cover("indiegames"), 1580, 890000, 45000, `["gaming","indie"]`, "Gaming", 2, defaultHLS},
		{4, 6, "video", "Lo-fi Beats to Study & Relax To - 3 Hour Mix", "Chill lo-fi hip hop beats perfect for studying, working, or relaxing. Enjoy 3 hours of uninterrupted vibes.", cover("lofi"), 10800, 4200000, 180000, `["music","lofi","study"]`, "Music", 14, defaultHLS},
		{5, 7, "video", "Breaking: Major Tech Layoffs Across Silicon Valley", "Analysis of the latest wave of tech layoffs, what it means for the industry, and how affected workers can navigate this challenging time.", cover("technews"), 720, 1300000, 32000, `["news","tech"]`, "News", 1, defaultHLS},
		{6, 8, "video", "Champions League Final Highlights", "Full highlights from an incredible Champions League final. Goals, saves, and drama.", cover("football"), 960, 5600000, 210000, `["sports","football"]`, "Sports", 1, defaultHLS},
		{7, 2, "video", "Building a Full-Stack App with Next.js 15", "Complete tutorial covering Next.js 15 features including server actions, parallel routes, and the new caching model.", cover("nextjs"), 2400, 340000, 15000, `["nextjs","tutorial"]`, "Education", 5, defaultHLS},
		{8, 4, "video", "Movie Recap: The Best Films of 2024", "A look back at the most memorable films of the year, from blockbusters to art-house gems.", cover("movies"), 1800, 780000, 28000, `["entertainment","movies"]`, "Entertainment", 4, defaultHLS},
		{9, 6, "video", "Guitar Tutorial: Learn 5 Songs in 30 Minutes", "Beginner-friendly guitar lesson. Learn to play popular songs with simple chords and strumming patterns.", cover("guitar"), 1860, 450000, 19000, `["music","tutorial"]`, "Music", 8, defaultHLS},
		{10, 5, "video", "Minecraft Speedrun World Record Attempt", "Watch this insane speedrun attempt of Minecraft. Will we break the record?", cover("minecraft"), 1200, 2100000, 95000, `["gaming","minecraft","speedrun"]`, "Gaming", 3, defaultHLS},
		{11, 2, "video", "CSS Has Changed: Modern Techniques You Should Know", "Container queries, cascade layers, :has() selector, and more. CSS in 2024 is incredibly powerful.", cover("moderncss"), 1100, 280000, 12000, `["css","frontend"]`, "Tech", 7, defaultHLS},
		{12, 8, "video", "NBA Playoffs: Top 10 Plays of the Week", "The most jaw-dropping dunks, assists, and game-winners from this week in the NBA playoffs.", cover("nba"), 660, 3400000, 150000, `["sports","basketball"]`, "Sports", 2, defaultHLS},
	}

	// Shorts
	shorts := []videoSeed{
		{13, 2, "short", "30s Tip: Better Loading States", "Skeletons > spinners for content surfaces.", cover("short1"), 30, 99000, 6300, `["shorts"]`, "Tech", 1, ""},
		{14, 4, "short", "1 Minute CSS Grid Layout", "Practical grid recipe in under a minute.", cover("short2"), 45, 142000, 10800, `["css"]`, "Tech", 2, ""},
		{15, 8, "short", "Insane Goal from Last Night", "You have to see this bicycle kick!", cover("short3"), 15, 2800000, 180000, `["sports","football"]`, "Sports", 1, ""},
		{16, 5, "short", "Epic Gaming Clutch Moment", "1v5 clutch that broke the internet.", cover("short4"), 25, 1500000, 95000, `["gaming","clutch"]`, "Gaming", 3, ""},
	}

	// Replay
	replay := videoSeed{17, 4, "video", "Livestream Replay: Architecture AMA", "Replay of live architecture discussion. We covered microservices, monoliths, and everything in between.", cover("replay"), 3800, 42000, 1800, `["live","replay"]`, "Tech", 7, defaultHLS}

	allVideos := append(longVideos, shorts...)
	allVideos = append(allVideos, replay)

	for _, v := range allVideos {
		createdAt := now.Add(-time.Duration(v.daysAgo) * day)
		db.Exec(`INSERT INTO videos (id, channel_id, type, title, description, status, visibility,
			duration, view_count, like_count, dislike_count, comment_count, favorite_count,
			thumbnail_url, tags, category, hls_url, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, 'ready', 'public', ?, ?, ?, 0, 0, 0, ?, ?, ?, ?, ?, ?)`,
			v.id, v.channelID, v.vType, v.title, v.description,
			v.duration, v.views, v.likes,
			v.cover, v.tags, v.category, v.hlsURL, createdAt, createdAt)
	}
	db.Exec("SELECT setval('videos_id_seq', 100, true)")

	// === Live Rooms ===
	type liveSeed struct {
		id        uint64
		channelID uint64
		title     string
		cover     string
		viewers   int64
		category  string
		hoursAgo  float64
	}
	liveRooms := []liveSeed{
		{1, 4, "Building Creator Studio Live", cover("live1"), 12400, "Tech", 1},
		{2, 2, "Frontend System Design Review", cover("live2"), 5300, "Education", 1.44},
		{3, 5, "Late Night Gaming Session - Elden Ring DLC", cover("live3"), 28900, "Gaming", 2},
		{4, 6, "Live Music Production: Making a Beat from Scratch", cover("live4"), 8700, "Music", 0.5},
	}
	for _, l := range liveRooms {
		createdAt := now.Add(-time.Duration(l.hoursAgo * float64(time.Hour)))
		db.Exec(`INSERT INTO live_rooms (id, channel_id, title, description, status, stream_key,
			playback_url, viewer_count, peak_viewer_count, category, thumbnail_url, created_at, updated_at)
			VALUES (?, ?, ?, '', 'live', ?, ?, ?, ?, ?, ?, ?, ?)`,
			l.id, l.channelID, l.title,
			fmt.Sprintf("sk_%d", l.id), defaultHLS, l.viewers, l.viewers,
			l.category, l.cover, createdAt, createdAt)
	}
	db.Exec("SELECT setval('live_rooms_id_seq', 100, true)")

	// === Comments ===
	// Map frontend video IDs to DB IDs: v1=1, v3=3, v4=4, v6=6, v7=7, v10=10
	type commentSeed struct {
		id       uint64
		videoID  uint64
		userID   uint64
		content  string
		likes    int64
		parentID *uint64
		hoursAgo float64
	}
	parentC1 := uint64(1)
	parentC10 := uint64(10)
	commentSeeds := []commentSeed{
		{1, 1, 1, "This was the clearest explanation of query keys I've ever seen. Subscribed!", 120, nil, 2},
		{2, 1, 2, "Thanks! We will publish part 2 soon.", 32, &parentC1, 1},
		// c3 is a live chat message, stored as chat_message below
		{4, 1, 5, "Great video! I wish you covered more about React Server Components though.", 45, nil, 1.5},
		{5, 3, 1, "Number 7 was so good! I played it for like 40 hours straight.", 89, nil, 24},
		{6, 3, 6, "The soundtrack alone makes some of these worth playing.", 28, nil, 12},
		{7, 4, 1, "I listen to this every day while working. Thank you for making it!", 340, nil, 48},
		{8, 6, 5, "What a match! That last-minute goal was unreal.", 156, nil, 8},
		{9, 7, 1, "Finally a Next.js 15 tutorial that actually explains the caching changes properly.", 67, nil, 4},
		{10, 10, 8, "The RNG on that Ender Pearl was insane, no way that was legit lol", 203, nil, 10},
		{11, 10, 5, "It was 100% legit, I verified the seed. Check the description for proof.", 89, &parentC10, 8.9},
	}
	for _, c := range commentSeeds {
		createdAt := now.Add(-time.Duration(c.hoursAgo * float64(time.Hour)))
		db.Exec(`INSERT INTO comments (id, video_id, user_id, parent_id, content, like_count, reply_count,
			is_pinned, is_hearted, status, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, 0, false, false, 'active', ?, ?)`,
			c.id, c.videoID, c.userID, c.parentID, c.content, c.likes, createdAt, createdAt)
	}
	db.Exec("SELECT setval('comments_id_seq', 100, true)")

	// === Chat Messages (for live rooms) ===
	// c3 from mock: entityType='live', entityId='l1' (room_id=1), author=users[0] (user_id=1)
	chatCreatedAt := now.Add(-time.Duration(200000 * time.Millisecond))
	db.Exec(`INSERT INTO chat_messages (id, room_id, user_id, type, content, status, created_at, updated_at)
		VALUES (1, 1, 1, 'text', '@creator can you share the repo later?', 'active', ?, ?)`,
		chatCreatedAt, chatCreatedAt)
	db.Exec("SELECT setval('chat_messages_id_seq', 100, true)")

	// === Subscriptions (user u1 subscribes to u2 and u4's channels) ===
	db.Exec(`INSERT INTO subscriptions (subscriber_id, channel_id, created_at, updated_at)
		VALUES (1, 2, ?, ?)`, now, now)
	db.Exec(`INSERT INTO subscriptions (subscriber_id, channel_id, created_at, updated_at)
		VALUES (1, 4, ?, ?)`, now, now)
	db.Exec("SELECT setval('subscriptions_id_seq', 100, true)")

	// === Playlist ===
	db.Exec(`INSERT INTO playlists (id, user_id, title, visibility, video_count, created_at, updated_at)
		VALUES (1, 2, 'Frontend Architecture Series', 'public', 3, ?, ?)`, now, now)
	db.Exec(`INSERT INTO playlist_items (playlist_id, video_id, position, created_at, updated_at)
		VALUES (1, 1, 0, ?, ?)`, now, now) // v1
	db.Exec(`INSERT INTO playlist_items (playlist_id, video_id, position, created_at, updated_at)
		VALUES (1, 2, 1, ?, ?)`, now, now) // v2
	db.Exec(`INSERT INTO playlist_items (playlist_id, video_id, position, created_at, updated_at)
		VALUES (1, 17, 2, ?, ?)`, now, now) // r1 = video ID 17
	db.Exec("SELECT setval('playlists_id_seq', 100, true)")
	db.Exec("SELECT setval('playlist_items_id_seq', 100, true)")

	// === Reports ===
	db.Exec(`INSERT INTO reports (id, reporter_id, content_type, content_id, reason, description, status, created_at, updated_at)
		VALUES (1, 1, 'video', 2, 'Copyright', '', 'open', ?, ?)`, now, now)
	db.Exec(`INSERT INTO reports (id, reporter_id, content_type, content_id, reason, description, status, created_at, updated_at)
		VALUES (2, 1, 'comment', 3, 'Harassment', '', 'open', ?, ?)`, now, now)
	db.Exec("SELECT setval('reports_id_seq', 100, true)")

	log.Println("All seed data inserted successfully.")
}
