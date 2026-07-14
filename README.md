# Domino

Local dev setup for running `api-gateway`, `user-service`, and `lobby-service` directly with `go run` (no Tilt/Kubernetes required). There's also a Tilt-based workflow that runs the services in a local Kubernetes cluster instead — see [Running with Tilt](#running-with-tilt) below.

## Prerequisites

- Go 1.25+
- Docker (for Postgres and Redis)
- [goose](https://github.com/pressly/goose) for migrations (`go install github.com/pressly/goose/v3/cmd/goose@latest`)
- `jq` (optional, for pretty-printing curl output)

## 1. Start Postgres

`user-service` stores users in Postgres.

```bash
docker run --name domino-postgres \
  -e POSTGRES_USER=domino \
  -e POSTGRES_PASSWORD=domino \
  -e POSTGRES_DB=domino \
  -p 5433:5432 \
  -v domino-postgres-data:/var/lib/postgresql/data \
  --restart unless-stopped \
  -d postgres:16
```

Connection string: `postgres://domino:domino@localhost:5433/domino?sslmode=disable`

Run migrations (from repo root, using `sql/schemas/`):

```bash
export GOOSE_DRIVER=postgres
export GOOSE_DBSTRING="postgres://domino:domino@localhost:5433/domino?sslmode=disable"
goose -dir sql/schemas up
```

Connect with `psql` if you want to inspect the data:

```bash
docker exec -it domino-postgres psql "postgres://domino:domino@localhost:5433/domino?sslmode=disable"
```

## 2. Start Redis

`lobby-service` stores lobby state in Redis.

```bash
docker run --name domino-redis \
  -p 6379:6379 \
  --restart unless-stopped \
  -d redis:7
```

## 3. Build the services

```bash
go build ./services/api-gateway/...
go build ./services/user-service/...
go build ./services/lobby-service/...
```

## 4. Run the services

Each runs in its own terminal.

**Terminal 1 — user-service** (listens on `:9095`, connects to Postgres on `localhost:5433`):

```bash
POSTGRESQL_URI="postgres://domino:domino@localhost:5433/domino?sslmode=disable" \
  go run ./services/user-service/cmd
```

**Terminal 2 — lobby-service** (listens on `:9094`, connects to Redis on `localhost:6379`):

```bash
go run ./services/lobby-service/cmd
```

**Terminal 3 — api-gateway** (listens on `:8081`):

```bash
USER_SERVICE_URL=localhost:9095 \
LOBBY_SERVICE_URL=localhost:9094 \
  go run ./services/api-gateway
```

> `USER_SERVICE_URL` / `LOBBY_SERVICE_URL` default to `user-service:9095` / `lobby-service:9094`, which resolve via Kubernetes DNS when deployed through Tilt. Outside a cluster those hostnames don't resolve, so override them to `localhost` for local runs.

> All three services also try to export traces to Jaeger (`JAEGER_ENDPOINT`, default `http://jaeger:14268/api/traces`). If you're not running Jaeger locally, this is harmless — trace export just fails silently in the background.

## 5. Test with curl

**Terminal 4:**

Create a guest user:

```bash
curl -s -X POST localhost:8081/users | jq
```

Fetch a user by id:

```bash
curl -s localhost:8081/users/<USER_ID> | jq
```

Create a lobby:

```bash
curl -s -X POST localhost:8081/lobbies \
  -H "Content-Type: application/json" \
  -d '{"userID":"host-1"}' | jq
```

Grab the `id` field from the response (the `secretToken` is `"abcdef"` — hardcoded default for now, see `services/lobby-service/internal/service/lobby.go`).

Join the lobby, using that `id` and secret code `"abcdef"`:

```bash
curl -s -X POST localhost:8081/lobbies/<LOBBY_ID>/join \
  -H "Content-Type: application/json" \
  -d '{"userID":"player-2","secretCode":"abcdef"}' | jq
```

A successful join returns the lobby with `players` now containing the joined player.

Start the game once enough players have joined:

```bash
curl -s -X POST localhost:8081/lobbies/<LOBBY_ID>/start \
  -H "Content-Type: application/json" \
  -d '{"hostID":"host-1"}' | jq
```

## Regenerating protobuf code

If you change a `.proto` file under `proto/`:

```bash
make generate-proto
```

## Running with Tilt

As an alternative to running each service with `go run`, [Tilt](https://tilt.dev) builds and deploys everything into a local Kubernetes cluster (via `minikube`), with live-reload on code changes.

> **Note:** the `Tiltfile` at the repo root still has leftovers from an earlier template (`trip-service`, `driver-service`, `payment-service`) that aren't part of this project's current stack, and it doesn't yet wire up `lobby-service`. It needs cleanup before it fully matches the `go run` workflow above — treat this section as how to get Tilt running, not a guarantee every resource in it is meaningful yet.

### Prerequisites

- Docker Desktop, with the WSL2 integration enabled for your distro (Docker Desktop → Settings → Resources → WSL Integration)
- [minikube](https://minikube.sigs.k8s.io/docs/start/)
- [Tilt](https://docs.tilt.dev/install.html)
- `kubectl` (installed automatically as a minikube dependency, or separately)

### 1. Start Docker

In WSL, make sure the Docker daemon is reachable — with Docker Desktop's WSL integration this just means Docker Desktop is running on Windows. Check with:

```bash
docker info
```

### 2. Start minikube

`minikube` spins up a local, single-node Kubernetes cluster (not a VPC — a VPC is a cloud networking construct; minikube gives you an actual cluster to deploy into):

```bash
minikube start
```

Check it's up:

```bash
minikube status
kubectl get nodes
```

### 3. Run Tilt

From the repo root:

```bash
tilt up
```

This opens the Tilt web UI (default `http://localhost:10350`) showing build/deploy status for each resource. Tilt watches your source files and rebuilds/redeploys automatically on change.

To tear everything down:

```bash
tilt down
```

You can also stop the cluster when you're done:

```bash
minikube stop
```

#NEXT STEPS:

- Add RefreshTokens: 
  - Creates a redis/sql database with user_id <-> SessionToken. If SessionToken in header is Ok to userID, then refresh JWT automatically




