-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS teams
(
    team_name  VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS teams;
-- +goose StatementEnd
