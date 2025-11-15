-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS pull_requests
(
    pull_request_id   VARCHAR(255) PRIMARY KEY,
    pull_request_name VARCHAR(255) NOT NULL,
    author_id         VARCHAR(255) NOT NULL REFERENCES users (user_id),
    status            VARCHAR(50)  NOT NULL DEFAULT 'OPEN' CHECK (status IN ('OPEN', 'MERGED')),
    created_at        TIMESTAMP    NOT NULL DEFAULT NOW(),
    merged_at         TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS pull_requests;
-- +goose StatementEnd
