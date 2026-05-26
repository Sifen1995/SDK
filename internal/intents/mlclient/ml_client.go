package mlclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	eventsModel "skykin-platform/internal/events/model"
	"strings"
)

type MLClient struct {
	BaseURL string
}

type MLResponse struct {
	Intent          string  `json:"intent"`
	Confidence      float64 `json:"confidence"`
	RewardTriggered bool    `json:"reward_triggered"`
	Reward          *struct {
		RewardType string  `json:"reward_type"`
		Amount     float64 `json:"amount"`
		Currency   string  `json:"currency"`
		Message    string  `json:"message"`
	} `json:"reward"`
}

func NewMLClient(baseURL string) *MLClient {
	return &MLClient{
		BaseURL: strings.TrimSpace(baseURL),
	}
}

func (c *MLClient) predictEndpoint() string {
	base := strings.TrimSuffix(c.BaseURL, "/")
	for strings.HasSuffix(base, "/predict-intent") {
		base = strings.TrimSuffix(base, "/predict-intent")
	}
	return base + "/predict-intent"
}

func (c *MLClient) PredictIntent(userID string, events []eventsModel.Event) (*MLResponse, error) {
	mlEvents := make([]map[string]interface{}, len(events))
	for i, e := range events {
		meta := e.Metadata
		if meta == nil {
			meta = map[string]interface{}{}
		}
		mlEvents[i] = map[string]interface{}{
			"event_type": e.EventType,
			"metadata":   meta,
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
