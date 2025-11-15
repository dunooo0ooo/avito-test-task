-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users
(
    user_id    VARCHAR(255) PRIMARY KEY,
    username   VARCHAR(255) NOT NULL,
    team_name  VARCHAR(255) NOT NULL REFERENCES teams (team_name),
    is_active  BOOLEAN      NOT NULL DEFAULT true,
    created_at TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP    NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
