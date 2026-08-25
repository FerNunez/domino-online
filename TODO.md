# TODO

- Insert a stub `games` row on `GameStarted` (status `IN_PROGRESS`), filled in later by `handleGameEnded`'s `UpsertGame` — without it, abandoned games are invisible to `GetGamesByPlayerID`.
- Subscribe `history_consumer` to `RoundStarted` and insert a stub `rounds` row (status `IN_PROGRESS`) the same way, filled in later by `handleRoundOver`'s `UpsertRound`.
