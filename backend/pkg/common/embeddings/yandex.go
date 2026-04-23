package embeddings

import (
	"context"
	"errors"
	"fmt"
	"strings"

	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

const (
	DefaultModel        = ""
	DefaultBaseURL      = "https://ai.api.cloud.yandex.net/v1"
	DefaultMaxBatchSize = 128
)

type YandexAIConfig struct {
	ApiKey       string
	FolderID     string
	Model        string
	BaseURL      string
	MaxBatchSize int
}

type YandexAIEmbedder struct {
	client       *openai.Client
	model        string
	baseUrl      string
	maxBatchSize int
}

func NewYandexAIEmbedder(cfg YandexAIConfig) (*YandexAIEmbedder, error) {
	if strings.TrimSpace(cfg.ApiKey) == "" {
		return nil, errors.New("yandex api key is required")
	}

	if strings.TrimSpace(cfg.FolderID) == "" {
		return nil, errors.New("yandex folder id is required")
	}

	if strings.TrimSpace(cfg.Model) == "" {
		cfg.Model = DefaultModel
	}

	if strings.TrimSpace(cfg.BaseURL) == "" {
		cfg.BaseURL = DefaultBaseURL
	}

	if cfg.MaxBatchSize <= 0 {
		cfg.MaxBatchSize = DefaultMaxBatchSize
	}

	client := openai.NewClient(
		option.WithAPIKey(cfg.ApiKey),
		option.WithBaseURL(cfg.BaseURL),
	)

	return &YandexAIEmbedder{
		client:       &client,
		model:        cfg.Model,
		baseUrl:      cfg.BaseURL,
		maxBatchSize: cfg.MaxBatchSize,
	}, nil
}

func (e *YandexAIEmbedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	prepared := make([]string, 0, len(texts))
	indexMap := make([]int, 0, len(texts))

	for i, text := range texts {
		text = normalizeText(text)
		if text == "" {
			continue
		}
		prepared = append(prepared, text)
		indexMap = append(indexMap, i)
	}

	if len(prepared) == 0 {
		return nil, errors.New("no non-empty texts to embed")
	}

	result := make([][]float32, len(texts))
	for start := 0; start < len(prepared); start += e.maxBatchSize {
		end := start + e.maxBatchSize
		if end > len(prepared) {
			end = len(prepared)
		}

		batch := prepared[start:end]
		vectors, err := e.embedBatch(ctx, batch)
		if err != nil {
			return nil, err
		}

		for i, vec := range vectors {
			originalIdx := indexMap[start+i]
			result[originalIdx] = vec
		}
	}

	final := make([][]float32, 0, len(texts))
	for i := range result {
		if result[i] != nil {
			final = append(final, result[i])
		}
	}

	return final, nil
}

func (e *YandexAIEmbedder) EmbedOne(ctx context.Context, text string) ([]float32, error) {
	vectors, err := e.Embed(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	return vectors[0], nil
}

func (e *YandexAIEmbedder) embedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	params := openai.EmbeddingNewParams{
		Model: openai.EmbeddingModel(e.model),
		Input: openai.EmbeddingNewParamsInputUnion{
			OfArrayOfStrings: texts,
		},
		EncodingFormat: openai.EmbeddingNewParamsEncodingFormatFloat,
	}

	resp, err := e.client.Embeddings.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("create embeddings: %w", err)
	}

	if len(resp.Data) != len(texts) {
		return nil, fmt.Errorf("unexpected embeddings count: got=%d want=%d", len(resp.Data), len(texts))
	}

	vectors := make([][]float32, 0, len(resp.Data))
	for _, item := range resp.Data {
		vec := make([]float32, len(item.Embedding))
		for i, v := range item.Embedding {
			vec[i] = float32(v)
		}
		vectors = append(vectors, vec)
	}

	return vectors, nil
}

func normalizeText(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}

	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])
	}

	return strings.TrimSpace(strings.Join(lines, "\n"))
}
