package apiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

var ErrActiveDialog = errors.New("active dialog")

type StartMatchResponse struct {
	Route        string `json:"route"`
	Status       string `json:"status"`
	DialogID     *int64 `json:"dialog_id,omitempty"`
	QueueSize    *int64 `json:"queue_size,omitempty"`
	DisplayCount *int64 `json:"display_count,omitempty"`
	MatchRoute   string `json:"match_route"`
}

func (c *Client) StartMatch(ctx context.Context, telegramID int64) (StartMatchResponse, error) {
	return c.postMatch(ctx, "/match/start", map[string]any{"telegram_id": telegramID})
}

func (c *Client) CompleteMatch(ctx context.Context, telegramID int64, waitSec int) (StartMatchResponse, error) {
	return c.postMatch(ctx, "/match/complete", map[string]any{
		"telegram_id": telegramID,
		"wait_sec":    waitSec,
	})
}

func (c *Client) PollMatch(ctx context.Context, telegramID int64) (StartMatchResponse, error) {
	return c.postMatch(ctx, "/match/poll", map[string]any{"telegram_id": telegramID})
}

func (c *Client) CancelMatch(ctx context.Context, telegramID int64) error {
	body, err := json.Marshal(map[string]int64{"telegram_id": telegramID})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/match/cancel", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("api cancel match: status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}
	return nil
}

func (c *Client) postMatch(ctx context.Context, path string, payload any) (StartMatchResponse, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return StartMatchResponse{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+path, bytes.NewReader(body))
	if err != nil {
		return StartMatchResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return StartMatchResponse{}, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode == http.StatusConflict {
		return StartMatchResponse{}, ErrActiveDialog
	}
	if resp.StatusCode != http.StatusOK {
		return StartMatchResponse{}, fmt.Errorf("api match %s: status %d: %s", path, resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var result StartMatchResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return StartMatchResponse{}, err
	}
	return result, nil
}
