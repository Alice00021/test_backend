-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS operations
(
    id            SERIAL PRIMARY KEY,
    created_at    TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at    TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    name          VARCHAR(100),
    description   VARCHAR(100),
    average_time  INTEGER
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS operations;
-- +goose StatementEnd
