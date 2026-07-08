-- +goose Up
CREATE TABLE users(
    id UUID PRIMARY KEY,
    display_name text NOT NULL,
    type TEXT NOT NULL CHECK (type IN('guest'))
);

-- +goose Down
DROP TABLE users;
