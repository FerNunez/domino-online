# Game lifecycle — sequence diagram

Full flow from lobby creation to game finish, as actually wired today: lobby lifecycle is plain REST through the API gateway, gameplay is a JWT-gated per-lobby WebSocket, and services talk to each other over RabbitMQ. Redis is the state store for lobby/game aggregates, not a message transport.

```mermaid
sequenceDiagram
    autonumber
    actor P1 as Player 1 (Host)
    actor P2 as Player 2..4
    participant GW as API Gateway
    participant LS as Lobby Service
    participant GS as Game Service
    participant MQ as RabbitMQ
    participant R as Redis (state)

    P1->>GW: POST /auth/guest
    GW-->>P1: guest JWT

    P1->>GW: POST /lobbies
    GW->>LS: gRPC CreateLobby
    LS->>R: save LobbyModel (status=WAITING, host=slot 1)
    LS-->>GW: lobby
    GW-->>P1: { lobbyId, wsToken }

    par Player 2..4 join
        P2->>GW: POST /auth/guest
        GW-->>P2: guest JWT
        P2->>GW: POST /lobbies/{id}/join
        GW->>LS: gRPC JoinLobby
        LS->>R: append player slot
        LS-->>GW: updated lobby
        GW-->>P2: { lobbyId, wsToken }
    end

    par Every player connects
        P1->>GW: GET /lobbies/{id}/ws?wsToken=...
        GW->>GW: validate wsToken, register connection
        GW->>MQ: bind QueueConsumer (notify_lobby_queue, notify_game_queue)
    end

    P1->>GW: POST /lobbies/{id}/start
    GW->>LS: gRPC StartLobby
    LS->>LS: require host + full roster (4 players)
    LS->>R: status=IN_GAME
    LS->>MQ: publish lobby.cmd.game_start (GameStartCmd)

    MQ->>GS: consume lobby.cmd.game_start
    GS->>GS: NewGame: build+shuffle double-six set (28 tiles),\ndeal 7 tiles/player, starting player = holder of 6-6 double
    GS->>R: save GameModel
    GS->>MQ: publish game.game_started (player order, hand size, current turn)
    MQ->>GW: relay to lobby
    GW-->>P1: WS game.game_started
    GW-->>P2: WS game.game_started
    GS->>MQ: publish game.hand_dealt (per-player, targeted)
    MQ->>GW: relay (targeted)
    GW-->>P1: WS game.hand_dealt (own tiles)
    GW-->>P2: WS game.hand_dealt (own tiles)
    GS->>MQ: publish game.turn_changed
    MQ->>GW: relay
    GW-->>P1: WS game.turn_changed

    loop Until round ends
        alt Current player plays a tile
            P1->>GW: WS game.cmd.play_tile { tile, side }
            GW->>GS: gRPC PlayTile (synchronous)
            GS->>GS: validate turn, side, tile-in-hand, legality
            GS->>GS: place tile, remove from hand, reset pass streak
            alt hand now empty
                GS->>GS: status=ROUND_OVER, reason=domino
            else
                GS->>GS: advance CurrentTurn
            end
            GS->>R: save updated GameModel
            GS-->>GW: PlayTileResponse (incl. RoundResult if over)
            GW->>MQ: publish game.move_made
            MQ->>GW: relay to lobby
            GW-->>P1: WS game.move_made / turn update
            GW-->>P2: WS game.move_made / turn update
        else Current player has no legal move
            P1->>GW: WS game.cmd.pass
            GW->>GS: gRPC PassTurn
            GS->>GS: reject if a legal move exists (ErrHasLegalMove)
            GS->>GS: else increment PassStreak
            alt PassStreak >= 4
                GS->>GS: status=ROUND_OVER, reason=blocked
            else
                GS->>GS: advance CurrentTurn
            end
            GS->>R: save updated GameModel
            GS-->>GW: PassTurnResponse
            GW->>MQ: publish game.turn_changed
            MQ->>GW: relay to lobby
            GW-->>P1: WS game.turn_changed
        end
    end

    Note over GS: ResolveRoundResult()<br/>domino: empty-hand player wins, others scored by pip sum<br/>blocked: lowest pip-sum wins (tie = no winner)
    GW-->>P1: WS round/game result (from inline RoundResult)
    GW-->>P2: WS round/game result (from inline RoundResult)

    Note over GS,MQ: game-service also has an async consumer path<br/>(game.cmd.play_tile / game.cmd.pass via NotifyGame,<br/>publishing game.ended) that duplicates the gRPC path above.<br/>The gateway's WS handler currently calls gRPC directly,<br/>so this async path is unused/dead code today — not shown as live.
```

## Key references
- `services/api-gateway/http.go`, `services/api-gateway/ws.go`
- `services/lobby-service/internal/service/lobby.go`
- `services/lobby-service/internal/infrastructure/grpc/grpc_handler.go`
- `services/lobby-service/internal/infrastructure/events/lobby_publisher.go`
- `services/game-service/internal/domain/game.go`
- `services/game-service/internal/service/game.go`
- `services/game-service/internal/infrastructure/grpc/grpc_handler.go`
- `services/game-service/internal/infrastructure/events/game_consumer.go`
- `shared/contracts/events.go`
- `shared/messaging/{events.go,queue_consumer.go,connection_manager.go}`
