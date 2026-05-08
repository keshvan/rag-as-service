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
	DefaultModel        = "emb://%s/text-search-%s/latest"
	DefaultBaseURL      = "https://ai.api.cloud.yandex.net/v1"
	DefaultMaxBatchSize = 128
)

type YandexAIConfig struct {
	ApiKey       string
	FolderID     string
	BaseURL      string
	Model        string
	MaxBatchSize int
}

type YandexAIClient struct {
	client       *openai.Client
	folderID     string
	model        string
	maxBatchSize int
}

func NewYandexAIClient(cfg YandexAIConfig) (*YandexAIClient, error) {
	if strings.TrimSpace(cfg.ApiKey) == "" {
		return nil, errors.New("yandex api key is required")
	}

	if strings.TrimSpace(cfg.FolderID) == "" {
		return nil, errors.New("yandex folder id is required")
	}

	if strings.TrimSpace(cfg.BaseURL) == "" {
		cfg.BaseURL = DefaultBaseURL
	}

	if strings.TrimSpace(cfg.Model) == "" {
		cfg.Model = DefaultModel
	}

	if cfg.MaxBatchSize <= 0 {
		cfg.MaxBatchSize = DefaultMaxBatchSize
	}

	client := openai.NewClient(
		option.WithAPIKey(cfg.ApiKey),
		option.WithBaseURL(cfg.BaseURL),
		option.WithProject(cfg.FolderID),
	)

	return &YandexAIClient{
		client:       &client,
		folderID:     cfg.FolderID,
		model:        cfg.Model,
		maxBatchSize: cfg.MaxBatchSize,
	}, nil
}

func (e *YandexAIClient) Embed(ctx context.Context, texts []string, textType string) ([][]float32, error) {
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
		vectors, err := e.embedBatch(ctx, batch, textType)
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

func (e *YandexAIClient) EmbedOne(ctx context.Context, text string) ([]float32, error) {
	vectors, err := e.Embed(ctx, []string{text}, "query")
	if err != nil {
		return nil, err
	}
	return vectors[0], nil
}

func (e *YandexAIClient) Complete(ctx context.Context, model, systemPrompt, userPrompt string, temperature float64, maxTokens int64) (string, error) {
	if strings.TrimSpace(model) == "" {
		model = "yandexgpt-lite"
	}
	if !strings.HasPrefix(model, "gpt://") {
		model = fmt.Sprintf("gpt://%s/%s", e.folderID, model)
	}

	resp, err := e.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: openai.ChatModel(model),
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage(userPrompt),
		},
		Temperature: openai.Float(temperature),
		MaxTokens:   openai.Int(maxTokens),
	})
	if err != nil {
		return "", fmt.Errorf("create chat completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", errors.New("no choices in completion response")
	}

	return resp.Choices[0].Message.Content, nil
}

func (e *YandexAIClient) embedBatch(ctx context.Context, texts []string, textType string) ([][]float32, error) {
	model := fmt.Sprintf(e.model, e.folderID, textType)
	vectors := make([][]float32, len(texts))

	for i, text := range texts {
		params := openai.EmbeddingNewParams{
			Model: openai.EmbeddingModel(model),
			Input: openai.EmbeddingNewParamsInputUnion{
				OfString: openai.String(text),
			},
			EncodingFormat: openai.EmbeddingNewParamsEncodingFormatFloat,
		}

		resp, err := e.client.Embeddings.New(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("create embeddings: %w", err)
		}

		if len(resp.Data) != 1 {
			return nil, fmt.Errorf("unexpected embeddings count for text %d: got=%d want=1", i, len(resp.Data))
		}

		vec := make([]float32, len(resp.Data[0].Embedding))
		for j, v := range resp.Data[0].Embedding {
			vec[j] = float32(v)
		}
		vectors[i] = vec
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
