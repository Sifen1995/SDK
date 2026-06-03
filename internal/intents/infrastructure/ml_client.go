package infrastructure

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"skykin-platform/internal/events/domain"
)

type MLClient struct {
	BaseURL string
}

type MLResponse struct {
	Intent          string   `json:"intent"`
	Confidence      float64  `json:"confidence"`
	RewardTriggered bool     `json:"reward_triggered"`
	TopSignals      []string `json:"top_signals"`
}

func NewMLClient(baseURL string) *MLClient {
	return &MLClient{BaseURL: strings.TrimSpace(baseURL)}
}

func (c *MLClient) predictEndpoint() string {
	base := strings.TrimSuffix(c.BaseURL, "/")
	for strings.HasSuffix(base, "/predict-intent") {
		base = strings.TrimSuffix(base, "/predict-intent")
	}
	return base + "/predict-intent"
}

// PredictIntent sends user events to the ML microservice for feature-based prediction.
func (c *MLClient) PredictIntent(userID string, events []domain.Event) (*MLResponse, error) {
	mlEvents := make([]map[string]interface{}, len(events))
	for i, e := range events {
		meta := e.Metadata
		if meta == nil {
			meta = map[string]any{}
		}
		mlEvents[i] = map[string]interface{}{
			"event_type":  string(e.EventType),
			"domain":      e.Domain,
			"screen_name": e.ScreenName,
			"metadata":    meta,
			"session_id":  e.SessionID,
			"created_at":  e.CreatedAt.Format(time.RFC3339),
		}
	}

	payload := map[string]interface{}{
		"user_id": userID,
		"events":  mlEvents,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal ML payload: %w", err)
	}

	req, err := http.NewRequest("POST", c.predictEndpoint(), bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("build ML request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ML service error: %d", resp.StatusCode)
	}

	var result MLResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode ML response: %w", err)
	}
	return &result, nil
}
