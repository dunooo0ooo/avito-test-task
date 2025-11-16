-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_users_team_name
    ON users (team_name);

CREATE INDEX IF NOT EXISTS idx_pull_requests_author_id
    ON pull_requests (author_id);

CREATE INDEX IF NOT EXISTS idx_pull_requests_status_created_at
    ON pull_requests (status, created_at);

CREATE INDEX IF NOT EXISTS idx_pr_reviewers_reviewer_id
    ON pr_reviewers (reviewer_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_pr_reviewers_reviewer_id;
DROP INDEX IF EXISTS idx_pull_requests_status_created_at;
DROP INDEX IF EXISTS idx_pull_requests_author_id;
DROP INDEX IF EXISTS idx_users_team_name;
-- +goose StatementEnd
