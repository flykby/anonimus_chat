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
	Route      string `json:"route"`
	Status     string `json:"status"`
	DialogID   *int64 `json:"dialog_id,omitempty"`
	QueueSize  *int64 `json:"queue_size,omitempty"`
	MatchRoute string `json:"match_route"`
}

func (c *Client) StartMatch(ctx context.Context, telegramID int64) (StartMatchResponse, error) {
	body, err := json.Marshal(map[string]int64{"telegram_id": telegramID})
	if err != nil {
		return StartMatchResponse{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/match/start", bytes.NewReader(body))
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
		return StartMatchResponse{}, fmt.Errorf("api start match: status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var result StartMatchResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return StartMatchResponse{}, err
	}
	return result, nil
}
