-- +goose Up
CREATE TABLE games (
    game_id UUID PRIMARY KEY,
    lobby_id UUID NOT NULL,
    final_scores JSONB NOT NULL,
    team_winner TEXT NOT NULL,
    game_state TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE rounds (
    round_id UUID PRIMARY KEY,
    game_id UUID NOT NULL,
    round_number INT NOT NULL,
    starting_player_id TEXT NOT NULL,
    player_order TEXT[] NOT NULL,
    winner_team_id TEXT NOT NULL,
    reason TEXT NOT NULL,
    scores JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_rounds_game_id ON rounds(game_id);

CREATE TABLE actions (
    round_id UUID NOT NULL,
    action_number INT NOT NULL,
    player_id TEXT NOT NULL,
    action_type TEXT NOT NULL CHECK (action_type IN ('play', 'pass')),
    tile_left INT,
    tile_right INT,
    side TEXT,
    resulting_left_end INT,
    resulting_right_end INT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (round_id, action_number)
);

-- +goose Down
DROP TABLE actions;
DROP TABLE rounds;
DROP TABLE games;
