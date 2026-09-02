-- +goose Up

-- Total accepted plays/passes for the round, set once at RoundOver time.
-- Lets readers verify a round's actions have fully landed 
ALTER TABLE rounds ADD COLUMN action_count INT NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE rounds DROP COLUMN action_count;
