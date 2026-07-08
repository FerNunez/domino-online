# Domino

Local dev setup for running `api-gateway`, `user-service`, and `lobby-service` directly with `go run` (no Tilt/Kubernetes required).

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
