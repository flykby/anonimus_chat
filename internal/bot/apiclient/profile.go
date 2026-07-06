package apiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type ProfileView struct {
	PublicUUID       string     `json:"public_uuid"`
	Age              int16      `json:"age"`
	Gender           string     `json:"gender"`
	Seeking          string     `json:"seeking"`
	Language         string     `json:"language"`
	PremiumActive    bool       `json:"premium_active"`
	PremiumExpiresAt *time.Time `json:"premium_expires_at,omitempty"`
}

func (c *Client) GetProfileView(ctx context.Context, telegramID int64) (ProfileView, error) {
	url := fmt.Sprintf("%s/users/by-telegram/%d/profile", c.BaseURL, telegramID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return ProfileView{}, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return ProfileView{}, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode == http.StatusNotFound {
		return ProfileView{}, ErrNotRegistered
	}
	if resp.StatusCode != http.StatusOK {
		return ProfileView{}, fmt.Errorf("api profile view: status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var view ProfileView
	if err := json.Unmarshal(body, &view); err != nil {
		return ProfileView{}, err
	}
	return view, nil
}
