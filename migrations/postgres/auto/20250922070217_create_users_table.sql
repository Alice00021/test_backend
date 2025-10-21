-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users
(
    id           SERIAL PRIMARY KEY,
    created_at   TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at   TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    name         VARCHAR(100),
    surname      VARCHAR(100),
    username     VARCHAR(100),
    email        VARCHAR(100) UNIQUE NOT NULL,
    password     VARCHAR(100) UNIQUE NOT NULL,
    role         VARCHAR(100) NOT NULL,
    file_path    VARCHAR(100),
    verify_token VARCHAR(100),
    is_verified  BOOLEAN NOT NULL DEFAULT FALSE,
    rating       DOUBLE PRECISION NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
