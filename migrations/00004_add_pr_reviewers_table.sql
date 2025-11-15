-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS pr_reviewers
(
    pull_request_id VARCHAR(255) NOT NULL REFERENCES pull_requests (pull_request_id),
    reviewer_id     VARCHAR(255) NOT NULL REFERENCES users (user_id),
    PRIMARY KEY (pull_request_id, reviewer_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS pr_reviewers;
-- +goose StatementEnd
