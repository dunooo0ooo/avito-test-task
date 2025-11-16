package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const baseURL = "http://localhost:8080"

func Test_FullFlow_Team_PR_Reviews_Stats(t *testing.T) {
	client := &http.Client{Timeout: 5 * time.Second}

	suffix := time.Now().UnixNano()
	teamName := fmt.Sprintf("team-int-%d", suffix)
	userAuthor := fmt.Sprintf("u-%d-1", suffix)
	userReviewer := fmt.Sprintf("u-%d-2", suffix)
	prID := fmt.Sprintf("pr-int-%d", suffix)

	teamReq := map[string]any{
		"team_name": teamName,
		"members": []map[string]any{
			{"user_id": userAuthor, "username": "Author", "is_active": true},
			{"user_id": userReviewer, "username": "Reviewer", "is_active": true},
		},
	}
	body, err := json.Marshal(teamReq)
	require.NoError(t, err)

	resp, err := client.Post(baseURL+"/team/add", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var teamResp struct {
		Team struct {
			TeamName string `json:"team_name"`
			Members  []struct {
				UserID   string `json:"user_id"`
				Username string `json:"username"`
				IsActive bool   `json:"is_active"`
			} `json:"members"`
		} `json:"team"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&teamResp))
	require.Equal(t, teamName, teamResp.Team.TeamName)
	require.Len(t, teamResp.Team.Members, 2)

	prReq := map[string]any{
		"pull_request_id":   prID,
		"pull_request_name": "Integration test PR",
		"author_id":         userAuthor,
	}
	body, err = json.Marshal(prReq)
	require.NoError(t, err)

	resp, err = client.Post(baseURL+"/pullRequest/create", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var prCreateResp struct {
		PR struct {
			PullRequestID     string   `json:"pull_request_id"`
			PullRequestName   string   `json:"pull_request_name"`
			AuthorID          string   `json:"author_id"`
			Status            string   `json:"status"`
			AssignedReviewers []string `json:"assigned_reviewers"`
		} `json:"pr"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&prCreateResp))
	require.Equal(t, prID, prCreateResp.PR.PullRequestID)
	require.Equal(t, userAuthor, prCreateResp.PR.AuthorID)
	require.Equal(t, "OPEN", prCreateResp.PR.Status)
	require.Contains(t, prCreateResp.PR.AssignedReviewers, userReviewer)

	resp, err = client.Get(baseURL + "/users/getReview?user_id=" + userReviewer)
	require.NoError(t, err)
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var reviewsResp struct {
		UserID       string `json:"user_id"`
		PullRequests []struct {
			PullRequestID   string `json:"pull_request_id"`
			PullRequestName string `json:"pull_request_name"`
			AuthorID        string `json:"author_id"`
			Status          string `json:"status"`
		} `json:"pull_requests"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&reviewsResp))

	require.Equal(t, userReviewer, reviewsResp.UserID)
	require.NotEmpty(t, reviewsResp.PullRequests)

	found := false
	for _, pr := range reviewsResp.PullRequests {
		if pr.PullRequestID == prID {
			found = true
			assert.Equal(t, "OPEN", pr.Status)
			break
		}
	}
	assert.True(t, found, "created PR must be in review list for reviewer")

	resp, err = client.Get(baseURL + "/stats/reviewers")
	require.NoError(t, err)
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var statsResp struct {
		Stats []struct {
			UserID string `json:"user_id"`
			Count  int64  `json:"count"`
		} `json:"stats"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&statsResp))

	hasReviewer := false
	for _, s := range statsResp.Stats {
		if s.UserID == userReviewer && s.Count > 0 {
			hasReviewer = true
			break
		}
	}
	assert.True(t, hasReviewer, "stats must contain reviewer with count > 0")
}
