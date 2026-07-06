## Manual Test: Lobby create/join

### Recompile

```bash
go build ./services/api-gateway/...
go build ./services/lobby-service/...
```

### Run api-gateway + lobby-service

Terminal 1 — lobby-service:
```bash
go run ./services/lobby-service/cmd
```
You should see `Lobby service listening on :9094`.

Terminal 2 — api-gateway:
```bash
LOBBY_SERVICE_URL=localhost:9094 go run ./services/api-gateway
```
You should see `API gateway listening on :8081`.

> `LOBBY_SERVICE_URL=localhost:9094` is set manually here for local testing. In production/Tilt, this defaults to `lobby-service:9094`, which resolves via Kubernetes' internal DNS (CoreDNS) once the service is deployed there. Outside a cluster that hostname doesn't resolve to anything, so it must be overridden to `localhost` when running the two processes directly.

### Test with curl

Terminal 3:

Create a lobby:
```bash
curl -s -X POST localhost:8081/lobbies \
  -H "Content-Type: application/json" \
  -d '{"userID":"host-1"}' | jq
```
Grab the `id` field from the response (the `secretToken` will be `"abcdef"` — hardcoded default for now, see `services/lobby-service/internal/service/service.go`).

Join it, using that `id` and secret code `"abcdef"`:
```bash
curl -s -X POST localhost:8081/lobbies/<LOBBY_ID>/join \
  -H "Content-Type: application/json" \
  -d '{"userID":"player-2","secretCode":"abcdef"}' | jq
```

A successful join returns the lobby with `players` now containing the joined player.
