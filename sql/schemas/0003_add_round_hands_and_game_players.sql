-- +goose Up

-- Initial dealt hand per player per round.
CREATE TABLE round_hands (
    round_id UUID NOT NULL,
    player_id TEXT NOT NULL,
    tiles JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (round_id, player_id)
);

-- Indexed join between a player and the games they took part in
CREATE TABLE game_players (
    game_id UUID NOT NULL,
    player_id TEXT NOT NULL,
    team_id TEXT NOT NULL,
    PRIMARY KEY (game_id, player_id)
);
CREATE INDEX idx_game_players_player_id ON game_players(player_id);

-- +goose Down
DROP TABLE game_players;
DROP TABLE round_hands;
