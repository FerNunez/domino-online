-- name: UpsertGame :exec
INSERT INTO games (game_id, lobby_id, final_scores, team_winner, game_state)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (game_id) DO UPDATE
SET lobby_id = $2, final_scores = $3, team_winner = $4, game_state = $5;

-- name: UpsertRound :exec
INSERT INTO rounds (round_id, game_id, round_number, starting_player_id, player_order, winner_team_id, reason, scores, action_count)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (round_id) DO UPDATE
SET game_id = $2, round_number = $3, starting_player_id = $4, player_order = $5, winner_team_id = $6, reason = $7, scores = $8, action_count = $9;

-- name: InsertAction :exec
INSERT INTO actions (round_id, action_number, player_id, action_type, tile_left, tile_right, side, resulting_left_end, resulting_right_end)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (round_id, action_number) DO NOTHING;

-- name: GetActionsByRoundID :many
SELECT * FROM actions
WHERE round_id = $1
ORDER BY action_number;

-- name: GetRoundsByGameID :many
SELECT * FROM rounds
WHERE game_id = $1
ORDER BY round_number;

-- name: GetRoundByID :one
SELECT * FROM rounds
WHERE round_id = $1;

-- name: UpsertHand :exec
INSERT INTO round_hands (round_id, player_id, tiles)
VALUES ($1, $2, $3)
ON CONFLICT (round_id, player_id) DO NOTHING;

-- name: GetHandsByRoundID :many
SELECT * FROM round_hands
WHERE round_id = $1;

-- name: UpsertGamePlayer :exec
INSERT INTO game_players (game_id, player_id, team_id)
VALUES ($1, $2, $3)
ON CONFLICT (game_id, player_id) DO NOTHING;

-- name: GetGamesByPlayerID :many
SELECT games.* FROM games
JOIN game_players ON game_players.game_id = games.game_id
WHERE game_players.player_id = $1
ORDER BY games.created_at DESC
LIMIT $2 OFFSET $3;
