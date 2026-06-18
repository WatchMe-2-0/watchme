# 🎬 WATCHME

**Self-hosted movie streaming platform for your local network.**

Drop a magnet link, WATCHME downloads the movie and streams it to any device on your network — no cloud, no subscriptions.

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.24-00ADD8?style=flat&logo=go" />
  <img src="https://img.shields.io/badge/Vite-6-646CFF?style=flat&logo=vite" />
  <img src="https://img.shields.io/badge/PostgreSQL-16-336791?style=flat&logo=postgresql" />
  <img src="https://img.shields.io/badge/Docker-Compose-2496ED?style=flat&logo=docker" />
</p>

---

## Features

- 🧲 **Magnet Link Downloads** — Paste a magnet link, movie downloads and becomes streamable
- 📡 **Local Network Streaming** — Any device on your WiFi can watch (phone, TV, tablet)
- 🔐 **Multi-Profile Auth** — Netflix-style profiles with PIN codes and kids mode
- 🎬 **TMDB Integration** — Browse trending, top-rated, search movies with posters
- 🌌 **GPU Aurora Shader** — GLSL-powered Icelandic aurora borealis background
- 🎨 **3 Themes** — Tron Blue (default), Night Mode, Dracula
- 🔒 **Admin Controls** — Manage profiles, settings, recovery key
- 📊 **Real-time Downloads** — SSE progress with speed, ETA, peer count
- 🐳 **Docker Ready** — Single `docker compose up` to run everything

---

## Quick Start

### Docker (Recommended)

```bash
git clone https://github.com/WatchMe-2-0/watchme.git
cd watchme
docker compose up -d
```

Open `http://localhost:3000` in your browser.

### Manual (Development)

**Prerequisites:** Go 1.24+, Node.js 22+, PostgreSQL 16

```bash
# Backend
cd backend
cp secrets/example.env .env  # Configure database
go mod tidy
go run .

# Frontend (new terminal)
cd frontend
npm install
npm run dev
```

- Frontend: `http://localhost:3000`
- Backend API: `http://localhost:8000`

---

## Architecture

```
watchme/
├── backend/           # Go (Fiber) API server
│   ├── auth/          # JWT, middleware, profile handlers
│   ├── config/        # JSON config, DB connection
│   ├── handlers/      # Movie, download, TMDB, settings
│   ├── models/        # GORM models
│   ├── torrent/       # anacrolix/torrent engine + worker pool
│   ├── tmdb/          # TMDB API client + LRU cache
│   ├── middleware/     # CORS, logger
│   └── utils/         # Response helpers, JWT secret gen
├── frontend/          # Vite + Vanilla JS SPA
│   ├── src/aurora/    # GLSL aurora shader (Three.js)
│   ├── src/pages/     # Landing, Setup, Dashboard, Browse...
│   ├── src/router/    # Hash-based SPA router
│   ├── src/state/     # Reactive pub/sub store
│   └── src/styles/    # CSS design system + themes
└── docker-compose.yml # PostgreSQL + Backend + Frontend
```

---

## API Endpoints

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/health` | — | Health check |
| GET | `/api/auth/status` | — | Check setup status |
| POST | `/api/auth/setup` | — | Create admin account |
| POST | `/api/auth/login` | — | PIN-based login |
| POST | `/api/auth/logout` | ✓ | Logout |
| GET | `/api/profiles/` | — | List profiles |
| POST | `/api/profiles/` | Admin | Create profile |
| GET | `/api/movies` | ✓ | List movies |
| POST | `/api/upload` | ✓ | Upload movie file |
| GET | `/api/stream/:id` | ✓ | Stream movie (range requests) |
| DELETE | `/api/movies/:id` | ✓ | Delete movie |
| POST | `/api/downloads` | ✓ | Start torrent download |
| GET | `/api/downloads/progress` | ✓ | SSE download progress |
| GET | `/api/tmdb/trending` | ✓ | TMDB trending movies |
| GET | `/api/tmdb/search` | ✓ | Search TMDB |

---

## Configuration

After first setup, configure in Settings:

- **TMDB API Key** — Get one free at [themoviedb.org](https://www.themoviedb.org/settings/api)
- **Download Directory** — Where movies are stored
- **Max Concurrent Downloads** — Parallel torrent slots (default: 3)

---

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.24, Fiber, GORM |
| Database | PostgreSQL 16 |
| Torrent | anacrolix/torrent |
| Frontend | Vite, Vanilla JS, Three.js |
| Styling | CSS (glassmorphism, GLSL shaders) |
| Auth | JWT (HS256), bcrypt, httpOnly cookies |
| Deployment | Docker Compose |

---

## License

MIT — Do what you want, just don't blame us.
