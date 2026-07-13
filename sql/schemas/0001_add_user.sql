-- +goose Up
CREATE TABLE users(
    id UUID PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    hashed_password TEXT UNIQUE NOT NULL,
    display_name TEXT NOT NULL
);

-- +goose Down
DROP TABLE users;
