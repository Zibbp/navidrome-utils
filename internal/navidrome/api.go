// Package navidrome implements a client for the Navidrome API.
package navidrome

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is used to interact with the Navidrome API.
type Client struct {
	// BaseURL is the API server URL
	BaseURL string
	// HTTPClient is the underlying HTTP client used to make requests.
	HTTPClient *http.Client
	// RetryCount is the number of retries to attempt for failed requests.
	RetryCount int
	// RetryWait is the wait duration between retries.
	RetryWait time.Duration
	// AuthToken holds the JWT token received from the Login endpoint.
	AuthToken string
}

// NewClient returns a new Navidrome API client.
// If httpClient is nil, a default client with a 30-second timeout is used.
func NewClient(baseURL string, retryCount int, retryWait time.Duration, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	return &Client{
		BaseURL:    baseURL,
		HTTPClient: httpClient,
		RetryCount: retryCount,
		RetryWait:  retryWait,
	}
}

// -----------------------------
// Request and Response Models
// -----------------------------

// LoginRequest is the request body for /auth/login.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse is assumed to be the response body from /auth/login.
type LoginResponse struct {
	Token string `json:"token"`
}

// CreatePlaylistRequest is the request body for creating or updating a playlist.
type CreatePlaylistRequest struct {
	Comment string `json:"comment"`
	Name    string `json:"name"`
	Public  bool   `json:"public"`
}

// CreatePlaylistResponse represents the response from the CreatePlaylist or UpdatePlaylist endpoint.
type CreatePlaylistResponse struct {
	ID string `json:"id"`
}

// AddTrackToPlaylistRequest is the request body for adding tracks to a playlist.
type AddTrackToPlaylistRequest struct {
	IDs []string `json:"ids"`
}

// Playlist represents a playlist resource.
// (Adjust fields as needed; the spec does not define the full response schema.)
type Playlist struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Comment string   `json:"comment"`
	Public  bool     `json:"public"`
	Tracks  []string `json:"tracks,omitempty"`
}

// -----------------------------
// API Methods
// -----------------------------

// Login authenticates with the API using the given username and password.
// It sends a POST to /auth/login and, on success, saves the JWT token for
// use in subsequent requests.
func (c *Client) Login(ctx context.Context, username, password string) error {
	url := fmt.Sprintf("%s/auth/login", c.BaseURL)
	// Prepare the request body.
	loginReq := LoginRequest{
		Username: username,
		Password: password,
	}
	reqBody, err := json.Marshal(loginReq)
	if err != nil {
		return err
	}
	// Use bytes.NewReader so that the request body is rewindable.
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.doRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("login failed: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	// Decode the login response.
	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return err
	}

	c.AuthToken = loginResp.Token
	return nil
}

// CreatePlaylist sends a POST request to /api/playlist to create a new playlist.
// It returns the playlist ID on success.
func (c *Client) CreatePlaylist(ctx context.Context, reqData *CreatePlaylistRequest) (string, error) {
	url := fmt.Sprintf("%s/api/playlist", c.BaseURL)
	reqBody, err := json.Marshal(reqData)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	c.setAuthHeader(req)

	resp, err := c.doRequest(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("create playlist failed: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var cpResp CreatePlaylistResponse
	if err := json.NewDecoder(resp.Body).Decode(&cpResp); err != nil {
		return "", err
	}
	return cpResp.ID, nil
}

// UpdatePlaylist sends a PUT request to /api/playlist/{playlistID} to update an existing playlist.
// It takes the same request body as CreatePlaylist and returns the updated playlist ID on success.
func (c *Client) UpdatePlaylist(ctx context.Context, playlistID string, reqData *CreatePlaylistRequest) (string, error) {
	url := fmt.Sprintf("%s/api/playlist/%s", c.BaseURL, playlistID)
	reqBody, err := json.Marshal(reqData)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(reqBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	c.setAuthHeader(req)

	resp, err := c.doRequest(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("update playlist failed: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var upResp CreatePlaylistResponse
	if err := json.NewDecoder(resp.Body).Decode(&upResp); err != nil {
		return "", err
	}
	return upResp.ID, nil
}

// GetPlaylists sends a GET request to /api/playlist to retrieve all playlists.
func (c *Client) GetPlaylists(ctx context.Context) ([]Playlist, error) {
	url := fmt.Sprintf("%s/api/playlist", c.BaseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	c.setAuthHeader(req)

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get playlists failed: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var playlists []Playlist
	if err := json.NewDecoder(resp.Body).Decode(&playlists); err != nil {
		return nil, err
	}
	return playlists, nil
}

// AddTrackToPlaylist sends a POST request to /api/playlist/{playlistID}/tracks to add tracks.
func (c *Client) AddTrackToPlaylist(ctx context.Context, playlistID string, reqData *AddTrackToPlaylistRequest) error {
	url := fmt.Sprintf("%s/api/playlist/%s/tracks", c.BaseURL, playlistID)
	reqBody, err := json.Marshal(reqData)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	c.setAuthHeader(req)

	resp, err := c.doRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("add track to playlist failed: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}
	return nil
}

// GetPlaylistByID sends a GET request to /api/playlist/{playlistID} to retrieve details.
func (c *Client) GetPlaylistByID(ctx context.Context, playlistID string) (*Playlist, error) {
	url := fmt.Sprintf("%s/api/playlist/%s", c.BaseURL, playlistID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	c.setAuthHeader(req)

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get playlist by id failed: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var playlist Playlist
	if err := json.NewDecoder(resp.Body).Decode(&playlist); err != nil {
		return nil, err
	}
	return &playlist, nil
}

// setAuthHeader adds the X-ND-Authorization header if an AuthToken is set.
func (c *Client) setAuthHeader(req *http.Request) {
	if c.AuthToken != "" {
		req.Header.Set("X-ND-Authorization", fmt.Sprintf("Bearer %s", c.AuthToken))
	}
}

// -----------------------------
// HTTP Request with Retry Logic
// -----------------------------

// doRequest executes the given HTTP request with retry logic.
// It retries on network errors or when the response status is in the 5xx range.
func (c *Client) doRequest(req *http.Request) (*http.Response, error) {
	var lastErr error

	// For requests with a non-nil body (that is rewindable, e.g. a bytes.Reader),
	// we can rely on req.GetBody to reset the body on each retry.
	// (http.NewRequest with bytes.NewReader sets GetBody automatically.)
	for attempt := 0; attempt <= c.RetryCount; attempt++ {
		// If this is a retry and the request body is non-nil, reset it.
		if attempt > 0 && req.Body != nil && req.GetBody != nil {
			bodyReader, err := req.GetBody()
			if err != nil {
				return nil, fmt.Errorf("failed to get request body on retry: %w", err)
			}
			req.Body = bodyReader
		}

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			lastErr = err
			// Wait before retrying, unless the context has been cancelled.
			if waitErr := sleepContext(req.Context(), c.RetryWait); waitErr != nil {
				return nil, waitErr
			}
			continue
		}

		// If the response status is 5xx, treat as a transient error.
		if resp.StatusCode >= http.StatusInternalServerError {
			// Drain and close the response body to reuse the connection.
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			lastErr = fmt.Errorf("server error: status %d", resp.StatusCode)
			if waitErr := sleepContext(req.Context(), c.RetryWait); waitErr != nil {
				return nil, waitErr
			}
			continue
		}

		// Success!
		return resp, nil
	}

	if lastErr == nil {
		lastErr = errors.New("request failed after retries")
	}
	return nil, lastErr
}

// sleepContext sleeps for the given duration or returns early if the context is done.
func sleepContext(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
