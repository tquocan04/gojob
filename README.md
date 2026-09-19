# GoJob — Concurrent Job Processing System

A production-inspired **job queue** written in Go, using **PostgreSQL as both the data store and the broker**. The API server accepts jobs via REST, and a separate worker process claims and processes them concurrently with retries, exponential backoff, and crash recovery.

Built to demonstrate real-world background-processing concerns: atomic concurrency, graceful shutdown, and failure handling — without introducing a heavyweight broker such as Redis or a message queue.

---

## Highlights

- **PostgreSQL as the job queue** — no Redis, no RabbitMQ. Atomic job claiming with `SELECT ... FOR UPDATE SKIP LOCKED`, so a queued job is claimed by only one worker at a time.
- **Concurrent worker pool** — configurable number of workers claim and execute jobs in parallel.
- **Retry with exponential backoff** — a failed job is rescheduled with `2^(attempts-1)` seconds of delay, up to `max_attempts`.
- **Crash recovery** — a dedicated recovery worker requeues jobs stuck in `processing` after a stale timeout, allowing jobs to be retried after worker crashes.
- **Graceful shutdown** — on `SIGTERM`/`SIGINT`, workers stop claiming, finish their in-flight job, persist the result, and exit cleanly.
- **Containerized** — one `docker compose up` starts Postgres, applies migrations, and runs the API server + worker.

## Architecture

```
┌────────────────┐   POST /jobs    ┌──────────────────────┐
│                │ ───────────────►│                      │
│   API Server   │   QUEUED         │      PostgreSQL      │
│   (:8080)      │◄─────────────── │   (store + broker)   │
│                │  GET /jobs/{id}  │                      │
└────────────────┘                  └──────────────────────┘
                                            ▲
                                    FOR UPDATE SKIP LOCKED
                                            │
                                  ┌─────────┴─────────┐
                                  │   Worker Pool (3) │──► execute job
                                  └─────────┬─────────┘        │
                                            │             retry/complete
                                            └► Recovery Worker (requeues stale jobs)
```

The worker loop: claim one job (`queued` + expired `available_at`) → mark `processing` → execute → write final status. If processing fails, the job is requeued with backoff or marked `failed` once attempts are exhausted.

> **Design goal:** demonstrate reliable concurrent job processing with PostgreSQL while keeping the system intentionally simple and broker-free.

## Tech Stack

| Layer      | Technology                          |
|------------|-------------------------------------|
| Language   | Go 1.26 (standard library `net/http`) |
| Database   | PostgreSQL 16                        |
| DB driver  | `jackc/pgx/v5` (connection pool)     |
| Migrations | `golang-migrate`                     |
| Container  | Docker multi-stage build + Compose   |

## Project Structure

```
├── cmd/
│   ├── server/main.go      # REST API (POST/GET jobs, /health)
│   └── worker/main.go      # worker pool + recovery worker
├── internal/
│   ├── database/           # pgx pool setup
│   ├── handlers/           # HTTP handlers & JSON response helpers
│   ├── services/           # business logic, validation, DTO mapping
│   ├── repositories/       # SQL (atomic claim, status updates)
│   ├── workers/            # worker pool, worker loop, recovery worker
│   ├── processors/         # mock job executors (email/report/image)
│   ├── migrations/         # SQL migration files
│   ├── models/             # Job model + statuses
│   ├── dtos/               # request/response DTOs
│   └── routes/             # route registration
├── Dockerfile              # multi-stage build (server & worker)
├── docker-compose.yml      # postgres + migrate + server + worker
└── Makefile                # migration + local-dev helpers
```

## Quick Start (Docker)

**Prerequisite:** Docker with Compose v2.

```bash
docker compose up --build
```

This starts, in order:

1. `postgres` — database with a healthcheck.
2. `migrate` — applies SQL migrations, then exits (blocks the app until done).
3. `server` — REST API on **http://localhost:8080**.
4. `worker` — 3 concurrent workers consuming the queue.

> On the first run, migration output `1/u jobs` confirms the `jobs` table was created.

### Try it

```bash
# Health check
curl http://localhost:8080/health

# Enqueue a job
curl -X POST http://localhost:8080/jobs \
  -H "Content-Type: application/json" \
  -d '{"type":"email","payload":{"to":"user@example.com"}}'

# Track its status (workers poll the queue every 60s)
curl http://localhost:8080/jobs/<job-id>
```

```
POST /jobs
→ {"status":"create successful","data":{"id":"6b4e36cd-...","status":"queued"}}

GET /jobs/<id>
→ {"status":"successful","data":{"id":"6b4e36cd-...","status":"completed"}}
```

### Stop

```bash
docker compose down        # stop everything (data persists in the pgdata volume)
docker compose down -v     # stop and delete the database volume
```

## Run Without Docker (local dev)

1. Start a PostgreSQL instance and create a DB (e.g. `gojob`).
2. Copy `.env.example` to `.env` and adjust the credentials.
3. Apply migrations: `make migration-up` (requires `golang-migrate`).
4. In two terminals:

```bash
go run ./cmd/server
go run ./cmd/worker
```

## API

| Method | Path          | Body                            | Response codes |
|--------|---------------|---------------------------------|----------------|
| GET    | `/health`     | —                               | 200            |
| POST   | `/jobs`       | `{"type":"email","payload":{...}}` | 201 or 400   |
| GET    | `/jobs/{id}`  | —                               | 200, 400, 404  |

**Supported job types** (all simulated with ~80% success to exercise the retry machinery):

| Type     | Description         | Example payload                    |
|----------|---------------------|------------------------------------|
| `email`  | mock email send     | `{"to":"user@example.com"}`        |
| `report` | mock report generation | `{"period":"2026-09"}`           |
| `image`  | mock image processing | `{"path":"/uploads/a.png"}`      |

## Configuration

All settings are read from environment variables. Defaults are defined in `docker-compose.yml` via `${VAR:-default}`.

| Variable       | Default       | Description                    |
|----------------|---------------|--------------------------------|
| `DB_HOST`      | `localhost`   | Database host                  |
| `DB_PORT`      | `5432`        | Database port                  |
| `DB_NAME`      | `gojob`       | Database name                  |
| `DB_USER`      | `postgres`    | Database user                  |
| `DB_PASSWORD`  | `admin123`    | Database password              |
| `DB_SSLMODE`   | `disable`     | SSL mode for the connection    |

Hardcoded tuning knobs (in code): worker count `3`, claim poll interval `60s`, recovery interval `10s`, stale-job timeout `30s`, max attempts `3`.

## How Concurrency & Reliability Work

1. **Atomic claim** — `ClaimQueuedJob` runs `SELECT ... FOR UPDATE SKIP LOCKED` inside a transaction: concurrent workers never claim the same job. JJS: attempts +1, `locked_at` set, status → `processing`.
2. **Backoff retry** — failure with attempts left schedules `available_at = now + 2^(attempts-1)s`; exhausted attempts → `failed`.
3. **Crash recovery** — the recovery worker requeues jobs that remain `processing` with a stale `locked_at`, allowing them to be claimed again after a worker crash. This provides at-least-once processing semantics.
4. **Graceful shutdown** — on `SIGTERM` the pool stops claiming, finishes the in-flight job (result persisted with a fresh context), then the recovery worker stops and the DB connection closes last.
