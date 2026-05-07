package qdrant

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	baseURL    *url.URL
	collection string
	apiKey     string
	vectorName string
	httpClient *http.Client
}

type Config struct {
	URL        string
	APIKey     string
	Collection string
	VectorName string
	Timeout    time.Duration
}

type ScoredPoint struct {
	ID      string
	Score   float32
	Payload map[string]any
}

func New(cfg Config) (*Client, error) {
	if strings.TrimSpace(cfg.URL) == "" {
		return nil, errors.New("qdrant url is required")
	}
	if strings.TrimSpace(cfg.Collection) == "" {
		return nil, errors.New("qdrant collection is required")
	}

	baseURL, err := url.Parse(strings.TrimRight(cfg.URL, "/"))
	if err != nil {
		return nil, fmt.Errorf("parse qdrant url: %w", err)
	}
	if baseURL.Scheme == "" || baseURL.Host == "" {
		return nil, fmt.Errorf("invalid qdrant url: %s", cfg.URL)
	}

	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	return &Client{
		baseURL:    baseURL,
		collection: cfg.Collection,
		apiKey:     cfg.APIKey,
		vectorName: strings.TrimSpace(cfg.VectorName),
		httpClient: &http.Client{Timeout: timeout},
	}, nil
}

func (c *Client) Search(ctx context.Context, vector []float32, orgID string, limit int, scoreThreshold *float32) ([]ScoredPoint, error) {
	if len(vector) == 0 {
		return nil, errors.New("query vector is empty")
	}
	if strings.TrimSpace(orgID) == "" {
		return nil, errors.New("organization_id is required for qdrant search")
	}
	if limit <= 0 {
		return nil, errors.New("limit must be positive")
	}

	reqBody := searchRequest{
		Vector:      c.vector(vector),
		Limit:       limit,
		WithPayload: true,
		WithVector:  false,
		Filter: filter{
			Must: []condition{{
				Key: "organization_id",
				Match: match{
					Value: orgID,
				},
			}},
		},
	}
	if scoreThreshold != nil {
		reqBody.ScoreThreshold = scoreThreshold
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal qdrant search request: %w", err)
	}

	endpoint := c.baseURL.ResolveReference(&url.URL{
		Path: fmt.Sprintf("/collections/%s/points/search", url.PathEscape(c.collection)),
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build qdrant search request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("api-key", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("qdrant search request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("qdrant search failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(msg)))
	}

	var decoded searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return nil, fmt.Errorf("decode qdrant search response: %w", err)
	}

	points := make([]ScoredPoint, 0, len(decoded.Result))
	for _, item := range decoded.Result {
		points = append(points, ScoredPoint{
			ID:      pointIDToString(item.ID),
			Score:   item.Score,
			Payload: item.Payload,
		})
	}

	return points, nil
}

func (c *Client) vector(values []float32) any {
	if c.vectorName == "" {
		return values
	}

	return namedVector{
		Name:   c.vectorName,
		Vector: values,
	}
}

type searchRequest struct {
	Vector         any      `json:"vector"`
	Filter         filter   `json:"filter"`
	Limit          int      `json:"limit"`
	WithPayload    bool     `json:"with_payload"`
	WithVector     bool     `json:"with_vector"`
	ScoreThreshold *float32 `json:"score_threshold,omitempty"`
}

type namedVector struct {
	Name   string    `json:"name"`
	Vector []float32 `json:"vector"`
}

type filter struct {
	Must []condition `json:"must"`
}

type condition struct {
	Key   string `json:"key"`
	Match match  `json:"match"`
}

type match struct {
	Value string `json:"value"`
}

type searchResponse struct {
	Result []searchPoint `json:"result"`
}

type searchPoint struct {
	ID      any            `json:"id"`
	Score   float32        `json:"score"`
	Payload map[string]any `json:"payload"`
}

func pointIDToString(id any) string {
	switch v := id.(type) {
	case string:
		return v
	case float64:
		return fmt.Sprintf("%.0f", v)
	default:
		raw, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprint(v)
		}
		return string(raw)
	}
}
