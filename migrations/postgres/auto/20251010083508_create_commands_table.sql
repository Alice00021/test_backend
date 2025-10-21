-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS commands
(
    id                SERIAL PRIMARY KEY,
    created_at        TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at        TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    name              VARCHAR(100),
    reagent           VARCHAR(100) NOT NULL,
    system_name       VARCHAR(100) UNIQUE NOT NULL,
    default_address   VARCHAR(100)  NOT NULL,
    average_time      INTEGER,
    volume_waste       INTEGER,
    volume_drive_fluid  INTEGER,
    volume_container   INTEGER
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS commands;
-- +goose StatementEnd

