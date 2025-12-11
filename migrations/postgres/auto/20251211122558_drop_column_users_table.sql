-- +goose Up
-- +goose StatementBegin
alter table users
    drop column IF EXISTS verify_token;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table users
    add column IF NOT EXISTS verify_token VARCHAR(100);
-- +goose StatementEnd
