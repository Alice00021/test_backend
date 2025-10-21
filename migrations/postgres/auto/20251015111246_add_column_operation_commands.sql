-- +goose Up
-- +goose StatementBegin
alter table operation_commands
    add column IF NOT EXISTS address VARCHAR(100);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table operation_commands
drop column IF EXISTS address;
-- +goose StatementEnd
