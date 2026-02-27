# AI Studies Blog

A tech-aesthetic blog for publishing notes on AI, machine learning, and technology.
Built with Go + Fiber, SQLite, server-side HTML templates, and an EasyMDE Markdown editor.

---

## Stack

| Layer        | Tech                                                        |
|--------------|-------------------------------------------------------------|
| Backend      | Go + Fiber v2                                               |
| Database     | SQLite via `modernc.org/sqlite` (pure Go, no CGO)          |
| Frontend     | Go HTML templates (server-side rendering)                   |
| Editor       | EasyMDE (vendored, offline-capable)                         |
| Deployment   | Docker (multi-stage, static binary)                         |

---

## Local Development

### Prerequisites

- Go 1.24+
- (Optional) Docker & Docker Compose

### Run directly

```bash
# 1. Copy and configure environment variables
cp .env.example .env
# Edit .env — APP_SECRET and IP_HASH_SECRET must be set in production

# 2. Create the first admin user
go run -mod=mod ./scripts/seed.go -username admin -password "YourPassword"

# 3. Start the server
go run -mod=mod ./cmd/server
# → http://localhost:3000
```

### Run with Docker Compose

```bash
cp .env.example .env
docker compose up --build
# → http://localhost:3000
```

---

## First Login

The admin panel lives at a non-obvious URL: **`/studio`**

```
http://localhost:3000/studio
```

Log in and you'll be redirected to the dashboard.

---

## Seed Script

Creates or updates the admin user in the SQLite database:

```bash
go run -mod=mod ./scripts/seed.go \
  -username admin \
  -password "YourStrongPassword" \
  -email "you@example.com"
```

Run this once before starting the server. Re-running it updates the password.

---

## Environment Variables

| Variable            | Default                | Description |
|---------------------|------------------------|-------------|
| `APP_ENV`           | `development`          | `development` or `production` |
| `APP_PORT`          | `3000`                 | HTTP port |
| `APP_SECRET`        | *(required in prod)*   | 32-byte secret for AES-256-GCM session encryption. Generate: `openssl rand -hex 32` |
| `IP_HASH_SECRET`    | *(required in prod)*   | Secret for SHA-256 IP hashing in analytics |
| `DB_PATH`           | `./data/blog.db`       | SQLite database path |
| `UPLOAD_DIR`        | `./web/static/uploads` | Uploaded media directory |
| `UPLOAD_MAX_MB`     | `20`                   | Max upload size (MB) |
| `SESSION_DURATION`  | `24h`                  | Session TTL |
| `RATE_LIMIT_LOGIN`  | `5`                    | Max login attempts per window |
| `RATE_LIMIT_WINDOW` | `15m`                  | Rate-limit sliding window |
| `CSP_MODE`          | `lenient`              | `lenient` (dev) or `strict` (prod) |

> In **production**, missing `APP_SECRET` or `IP_HASH_SECRET` causes a fatal error at startup.

---

## Security

| Concern               | Implementation |
|-----------------------|----------------|
| Password storage      | bcrypt (cost 12) |
| Session data          | AES-256-GCM encrypted, stored in SQLite |
| Session cookie        | `HttpOnly`, `Secure` (prod), `SameSite=Lax` |
| IP privacy            | SHA-256(ip + secret) — raw IPs never stored |
| Rate limiting         | Sliding-window in-memory limiter on `POST /studio/login` |
| Security headers      | `X-Content-Type-Options`, `X-Frame-Options: DENY`, `Referrer-Policy`, `Permissions-Policy`, CSP |
| HSTS                  | Added in production mode |
| File uploads          | MIME type whitelist + size cap |

### CSP modes

- **`lenient`** (default for dev): Relaxed — allows `unsafe-inline`, all media sources. Suitable for local dev.
- **`strict`** (production): `script-src 'self'` only. EasyMDE is vendored locally so no CDN is needed.

### Production checklist

- [ ] Set `APP_ENV=production`
- [ ] Set `CSP_MODE=strict`
- [ ] Set strong `APP_SECRET` — `openssl rand -hex 32`
- [ ] Set strong `IP_HASH_SECRET` — `openssl rand -hex 32`
- [ ] Put behind a TLS-terminating reverse proxy (nginx, Caddy)

---

## Running Tests

```bash
go test -mod=mod ./tests/... -v
```

Integration tests use an in-memory SQLite database — no `.env` or running server needed.

### What's tested

- Security headers on all responses
- Rate limiting (429 + `Retry-After`) on login endpoint
- CSP present on all public routes
- Login, logout, session invalidation, protected-route redirect
- Post CRUD: create, publish, update, delete, slug uniqueness

---

## Project Structure

```
blog-ai/
├── cmd/server/main.go             # Entry point
├── internal/
│   ├── config/                    # Env-based config
│   ├── database/migrations/       # 5 SQL migration files
│   ├── middleware/                 # security, ratelimit, auth, analytics
│   ├── handler/public/            # Home, Post, Category, Timeline
│   ├── handler/studio/            # Auth, Dashboard, Posts, Metrics
│   ├── service/                   # Business logic
│   ├── repository/                # SQL queries
│   └── model/                     # Data structs
├── web/templates/                 # Go HTML templates
├── web/static/css/                # public.css, studio.css
├── web/static/js/                 # EasyMDE (vendored), editor.js
├── web/static/uploads/            # User-uploaded media
├── tests/integration/             # Security, auth, post tests
├── scripts/seed.go                # Create first admin user
├── Dockerfile                     # Multi-stage, CGO_ENABLED=0
├── docker-compose.yml
└── .env.example
```

---

## Admin Panel (`/studio`)

| Section      | Features |
|--------------|----------|
| Dashboard    | Total views, today's views, published posts, top 5 posts, 30-day chart |
| All Posts    | Status badges, publish/unpublish/delete, editor link |
| Post Editor  | EasyMDE with live preview, image/video/audio upload |
| Metrics      | Full view counts per post, ranked table, daily view chart |

### Media uploads in editor

Click the **↑ upload button** in the toolbar. Supported:
- Images: JPEG, PNG, GIF, WebP
- Video: MP4, WebM
- Audio: MP3, OGG

---

## Deployment on Hostinger VPS

```bash
git clone <your-repo> blog-ai && cd blog-ai
cp .env.example .env
# Edit: APP_ENV=production, CSP_MODE=strict, strong secrets

go run -mod=mod ./scripts/seed.go -username admin -password "StrongPassword"

docker compose up -d --build
```

Point nginx/Caddy at `localhost:3000`.
