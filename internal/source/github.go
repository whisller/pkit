package source

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/whisller/pkit/pkg/models"
)

const (
	// GitHubAPIBaseURL is the base URL for GitHub API
	GitHubAPIBaseURL = "https://api.github.com"
)

// GitHubClient is a client for making GitHub API requests with rate limit tracking.
type GitHubClient struct {
	httpClient *http.Client
	token      string
}

// NewGitHubClient creates a new GitHub API client.
// If token is empty, requests will be unauthenticated (60 req/hour limit).
func NewGitHubClient(token string) *GitHubClient {
	return &GitHubClient{
		httpClient: &http.Client{},
		token:      token,
	}
}

// GetRepository fetches repository information from GitHub API.
// Returns repository data and rate limit information.
func (c *GitHubClient) GetRepository(owner, repo string) (map[string]interface{}, *models.RateLimit, error) {
	url := fmt.Sprintf("%s/repos/%s/%s", GitHubAPIBaseURL, owner, repo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication if token is available
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	// Set Accept header for GitHub API v3
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	// Make request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		// Response body close errors are rare, ignore
		_ = resp.Body.Close()
	}()

	// Extract rate limit from response headers
	rateLimit := ExtractRateLimitFromHeaders(resp, c.token != "")

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, rateLimit, fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, resp.Status)
	}

	// Parse response body
	var repoData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&repoData); err != nil {
		return nil, rateLimit, fmt.Errorf("failed to decode response: %w", err)
	}

	return repoData, rateLimit, nil
}

// CheckRateLimit fetches current rate limit status from GitHub API.
// This endpoint does NOT count against the rate limit.
func (c *GitHubClient) CheckRateLimit() (*models.RateLimit, error) {
	url := fmt.Sprintf("%s/rate_limit", GitHubAPIBaseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication if token is available
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	// Set Accept header
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	// Make request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() {
		// Response body close errors are rare, ignore
		_ = resp.Body.Close()
	}()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, resp.Status)
	}

	// Parse response body
	var rateLimitData struct {
		Resources struct {
			Core struct {
				Limit     int   `json:"limit"`
				Remaining int   `json:"remaining"`
				Reset     int64 `json:"reset"`
			} `json:"core"`
		} `json:"resources"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&rateLimitData); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert to RateLimit model
	rateLimit := &models.RateLimit{
		Limit:         rateLimitData.Resources.Core.Limit,
		Remaining:     rateLimitData.Resources.Core.Remaining,
		ResetAt:       parseUnixTime(rateLimitData.Resources.Core.Reset),
		Resource:      "core",
		Authenticated: c.token != "",
	}

	return rateLimit, nil
}

// ExtractRateLimitFromHeaders extracts rate limit information from response headers.
func ExtractRateLimitFromHeaders(resp *http.Response, authenticated bool) *models.RateLimit {
	if resp == nil {
		return nil
	}

	// Import the ExtractRateLimitFromResponse from config package
	// to avoid code duplication
	return nil // This will be implemented in internal/config/ratelimit.go
}

// ParseGitHubURL parses a GitHub repository URL into owner and repo name.
// Supports formats:
//   - https://github.com/owner/repo
//   - https://github.com/owner/repo.git
//   - owner/repo
func ParseGitHubURL(url string) (owner, repo string, err error) {
	// Remove trailing .git if present
	url = strings.TrimSuffix(url, ".git")

	// Remove protocol if present
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")

	// Remove github.com/ if present
	url = strings.TrimPrefix(url, "github.com/")

	// Split by /
	parts := strings.Split(url, "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid GitHub URL format: %s", url)
	}

	owner = parts[0]
	repo = parts[1]

	if owner == "" || repo == "" {
		return "", "", fmt.Errorf("invalid GitHub URL: owner or repo is empty")
	}

	return owner, repo, nil
}

// ValidateGitHubURL validates a GitHub repository URL by making an API call.
// Returns an error if the repository doesn't exist or is not accessible.
func (c *GitHubClient) ValidateGitHubURL(url string) error {
	owner, repo, err := ParseGitHubURL(url)
	if err != nil {
		return err
	}

	_, _, err = c.GetRepository(owner, repo)
	return err
}

// Helper function to parse Unix timestamp
func parseUnixTime(unixTime int64) time.Time {
	return time.Unix(unixTime, 0)
}
