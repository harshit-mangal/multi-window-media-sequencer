# Multi-Window Media Sequencer with Synchronized Playback

A full-stack real-time media sequencing system for multiple independent
display windows.

Each window maintains its own ordered playlist and continuously plays it
using a deterministic **5-hour (18,000-second) cycle**. A
synchronization command temporarily overrides all windows with a
selected media item for a configured duration, then each window resumes
its normal deterministic playback without modifying its stored playlist.

## Live Deployment

-   **Frontend:** https://multi-window-media-sequencer-1.onrender.com
-   **Backend:** https://multi-window-media-sequencer-2vjr.onrender.com
-   **Backend health:**
    https://multi-window-media-sequencer-2vjr.onrender.com/health
-   **Repository:**
    https://github.com/harshit-mangal/multi-window-media-sequencer

The frontend and backend are deployed as separate Render Web Services,
with PostgreSQL used for persistent state.

> Render's free web services can spin down after inactivity, so the
> first request after inactivity may have a cold-start delay.

## Features

### Independent multi-window playback

-   Four independently configured display windows.
-   Each window has its own ordered playlist.
-   Windows can have different media, ordering, and durations.
-   Normal playback continues independently.

### Deterministic 5-hour playback

The playback cycle is fixed at:

``` text
5 hours = 18,000 seconds
```

Playback state is derived from server time and playlist configuration
instead of being written to PostgreSQL every second.

Conceptually:

``` text
CycleNumber  = floor(CurrentUnixTimestamp / 18000)
CycleElapsed = CurrentUnixTimestamp mod 18000
LoopElapsed  = CycleElapsed mod TotalPlaylistDuration
```

The active media item is resolved from cumulative playlist durations.

### Temporary synchronized playback

A selected media item can be synchronized across all windows.

The backend: 1. Creates a synchronization event. 2. Calculates a future
`startAt` using a small network buffer. 3. Calculates `endAt`. 4.
Broadcasts the event through WebSocket. 5. Clients switch at the shared
server-defined `startAt`. 6. Clients leave the override at `endAt` and
return to normal playback.

Synchronization does not reorder, delete, or replace stored playlist
items.

### Refresh and late-client recovery

Clients can call:

``` text
GET /api/sync/current
```

to reconstruct an active synchronization event after a refresh or late
connection.

### Dynamic playlist updates

Playlist changes are persisted and broadcast through:

``` text
PLAYLIST_UPDATED
```

so connected displays can update without a full browser reload.

### Media support

The player supports: - Images - Videos - Explicit blank/intermission
items

### Persistent storage

PostgreSQL stores: - Windows - Media assets - Playlist items -
Synchronization events

Migrations and seed data are included.

## Architecture

``` mermaid
graph TD
    Browser[React Display / Admin UI]
    Browser -->|REST API| Backend[Go Backend]
    Browser -->|WebSocket| Backend
    Backend --> Router[HTTP Router]
    Backend --> WSHub[WebSocket Hub]
    Backend --> Playback[Deterministic Playback Service]
    Backend --> Sync[Synchronization Service]
    Router --> Services[Application Services]
    Sync --> Services
    Playback --> Services
    Services --> Repositories[Repository Layer]
    Repositories --> Pool[pgx Connection Pool]
    Pool --> PostgreSQL[(PostgreSQL)]
```

## Technology Stack

### Backend

-   Go 1.27.1+
-   PostgreSQL
-   pgx/v5
-   Gorilla WebSocket
-   godotenv
-   Go HTTP server
-   Handler / service / repository architecture

### Frontend

-   React 19
-   Vite 8
-   Tailwind CSS v4
-   Lucide Icons
-   REST API client
-   WebSocket client

### Infrastructure

-   Docker
-   Docker Compose
-   Nginx
-   Render
-   PostgreSQL

## Project Structure

``` text
multi-window-media-sequencer/
├── backend/
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── config/
│   │   ├── database/
│   │   ├── handlers/
│   │   ├── middleware/
│   │   ├── models/
│   │   ├── repository/
│   │   ├── services/
│   │   └── websocket/
│   ├── migrations/
│   ├── seed/
│   ├── tests/
│   ├── postman_collection.json
│   ├── Dockerfile
│   ├── go.mod
│   └── .env.example
├── frontend/
│   ├── src/
│   │   ├── assets/
│   │   ├── components/
│   │   ├── hooks/
│   │   ├── services/
│   │   ├── utils/
│   │   ├── App.jsx
│   │   └── main.jsx
│   ├── public/
│   ├── Dockerfile
│   ├── nginx.conf
│   ├── package.json
│   └── .env.example
├── docker-compose.yml
├── render.yaml
├── README.md
└── .gitignore
```

## Database Model

### `windows`

Stores display-window configuration.

``` text
id
name
created_at
updated_at
```

### `media`

Stores media assets.

``` text
id
name
type
url
duration
created_at
updated_at
```

Supported types:

``` text
image
video
blank
```

### `playlist_items`

Associates media with windows and preserves ordering.

``` text
id
window_id
media_id
position
created_at
updated_at
```

### `sync_events`

Stores temporary synchronization events.

``` text
id
media_id
duration
start_at
end_at
status
created_at
```

## REST API

### Windows

`GET /api/windows`\
Returns all configured windows and playlists.

`GET /api/windows/:id`\
Returns one window and its calculated playback state.

### Playlists

`POST /api/windows/:id/playlist`

``` json
{
  "mediaId": 4,
  "position": 2
}
```

`PUT /api/windows/:id/playlist/:itemId`

``` json
{
  "position": 0
}
```

`DELETE /api/windows/:id/playlist/:itemId`

### Media

`GET /api/media`

`POST /api/media`

``` json
{
  "name": "Example Image",
  "type": "image",
  "url": "https://example.com/image.jpg",
  "duration": 10
}
```

`DELETE /api/media/:id`

### Synchronization

`POST /api/sync`

``` json
{
  "mediaId": 2,
  "duration": 20
}
```

`GET /api/sync/current`\
Returns the active synchronization state, if any.

`GET /health`\
Backend health check.

## WebSocket Events

  -----------------------------------------------------------------------
  Event                   Direction               Purpose
  ----------------------- ----------------------- -----------------------
  `SYNC_START`            Server → Client         Starts a synchronized
                                                  temporary override

  `SYNC_END`              Server → Client         Ends the
                                                  synchronization
                                                  override

  `PLAYLIST_UPDATED`      Server → Client         Notifies clients of
                                                  playlist changes
  -----------------------------------------------------------------------

## Local Development

### Prerequisites

-   Go 1.27.1+
-   Node.js 20+
-   Docker Desktop
-   Docker Compose
-   PostgreSQL 14+ if running PostgreSQL manually

### Recommended: Docker Compose

From the repository root:

``` bash
docker compose up --build
```

Local services:

``` text
Frontend:   http://localhost:3000
Backend:    http://localhost:8080
PostgreSQL: localhost:5432
```

### Run backend manually

``` bash
cd backend
cp .env.example .env
go run cmd/server/main.go
```

The backend attempts to connect to PostgreSQL, run migrations, seed
initial data, start the WebSocket hub, and start the HTTP server.

### Run backend tests

``` bash
cd backend
go test -v ./tests/...
```

### Run frontend manually

``` bash
cd frontend
npm install
npm run dev
```

Vite normally serves the frontend at:

``` text
http://localhost:5173
```

## Environment Variables

### Backend

Example local configuration:

``` env
PORT=8080
DATABASE_URL=postgres://postgres:postgres@localhost:5432/mediasequencer?sslmode=disable
ALLOWED_ORIGINS=*
SYNC_BUFFER_SECONDS=2
CYCLE_DURATION_SECONDS=18000
```

Production database credentials must be supplied by the hosting platform
and must not be committed to Git.

### Frontend

Local:

``` env
VITE_API_URL=http://localhost:8080/api
VITE_WS_URL=ws://localhost:8080/ws
```

Production:

``` env
VITE_API_URL=https://multi-window-media-sequencer-2vjr.onrender.com/api
VITE_WS_URL=wss://multi-window-media-sequencer-2vjr.onrender.com/ws
```

The React application reads these values through Vite's
`import.meta.env` configuration.

## Production Deployment

The production deployment uses three logical components:

``` text
React + Nginx
      |
      | HTTPS / WSS
      v
Go Backend
      |
      v
PostgreSQL
```

### Frontend Render service

``` text
Language: Docker
Root Directory: frontend
Dockerfile Path: Dockerfile
```

Production variables:

``` text
VITE_API_URL=https://multi-window-media-sequencer-2vjr.onrender.com/api
VITE_WS_URL=wss://multi-window-media-sequencer-2vjr.onrender.com/ws
```

Nginx serves the compiled React SPA and provides SPA route fallback.

### Backend Render service

``` text
Language: Docker
Root Directory: backend
Dockerfile Path: Dockerfile
```

The backend receives its production `DATABASE_URL` from the Render
PostgreSQL service and uses the platform-provided `PORT`.

## Docker Architecture

### Backend

Multi-stage build:

``` text
Go build stage
      ↓
Compiled Linux binary
      ↓
Alpine runtime image
```

### Frontend

Multi-stage build:

``` text
Node build stage
      ↓
Vite production bundle
      ↓
Nginx runtime image
```

## Seed Data

The repository includes seed data for the initial demonstration
environment, including multiple windows, media assets, and
window-specific playlists.

## Postman

Import:

``` text
backend/postman_collection.json
```

into Postman to test the REST API.

Useful endpoints include:

``` text
GET    /api/windows
GET    /api/windows/:id
GET    /api/media
POST   /api/media
POST   /api/windows/:id/playlist
PUT    /api/windows/:id/playlist/:itemId
DELETE /api/windows/:id/playlist/:itemId
POST   /api/sync
GET    /api/sync/current
GET    /health
```

## Design Decisions

### Deterministic playback

Playback position is calculated from time and playlist configuration
rather than persisted every second. This reduces database write load and
allows clients to reconstruct state after reconnecting.

### Server-time synchronization

Synchronization uses server-generated timestamps rather than relying on
every client receiving an immediate "start now" message.

### Non-destructive synchronization

A synchronization event acts as a temporary display override. Stored
playlists remain unchanged.

### Persistent configuration

Window, media, playlist, and synchronization configuration is persisted
in PostgreSQL.

## Assumptions and Trade-offs

1.  The normal playback cycle is fixed at 18,000 seconds.
2.  Playback state is calculated instead of persisted every second.
3.  Synchronization uses a small future start buffer to account for
    network latency.
4.  Synchronization does not mutate underlying playlists.
5.  Media URLs are stored as references; the application does not host
    or transcode media files.
6.  Third-party media availability depends on the external host.
7.  The current Render free-tier deployment is intended for
    demonstration/evaluation rather than high-availability production
    workloads.

## Known Demo Consideration

Some seeded demonstration videos use external media URLs. Their
availability can depend on the third-party host, browser policies, or
continued availability of the sample asset.

For production use, media should ideally be hosted on controlled object
storage/CDN infrastructure.

## Assignment Evaluation Checklist

  -----------------------------------------------------------------------
  Requirement                         Implementation
  ----------------------------------- -----------------------------------
  Multiple display windows            Independent window configurations

  Individual playlists                PostgreSQL-backed playlist items

  5-hour cycle                        Deterministic 18,000-second
                                      playback

  Continuous playback                 Client playback + deterministic
                                      state calculation

  Images                              Supported

  Videos                              Supported

  Explicit blank media                Supported

  Dynamic playlist updates            REST + WebSocket

  Temporary synchronization           Server-timestamped override

  Playlist preservation               Non-destructive sync

  Refresh recovery                    `/api/sync/current`

  Persistent storage                  PostgreSQL

  REST API                            Implemented

  WebSocket                           Implemented

  Docker                              Backend and frontend Dockerfiles

  Local orchestration                 Docker Compose

  Cloud deployment                    Render

  API testing                         Postman collection
  -----------------------------------------------------------------------

## Project Status

**Live and deployed.**

-   Frontend: https://multi-window-media-sequencer-1.onrender.com
-   Backend: https://multi-window-media-sequencer-2vjr.onrender.com
-   Health:
    https://multi-window-media-sequencer-2vjr.onrender.com/health

## Author

**Harshit Mangal**\
B.Tech --- Information Technology

GitHub: https://github.com/harshit-mangal Repository:
https://github.com/harshit-mangal/multi-window-media-sequencer
