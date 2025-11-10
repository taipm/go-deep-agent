package agent

import (
	"context"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

const (
	// Default embedding models
	EmbeddingModelSmall  = "text-embedding-3-small" // 1536 dimensions
	EmbeddingModelLarge  = "text-embedding-3-large" // 3072 dimensions
	EmbeddingModelAda002 = "text-embedding-ada-002" // 1536 dimensions (legacy)
)

// OpenAIEmbedding implements EmbeddingProvider using OpenAI's embedding API
type OpenAIEmbedding struct {
	client *openai.Client
	model  string
	config *EmbeddingConfig
}

// NewOpenAIEmbedding creates a new OpenAI embedding provider
// Supported models:
//   - text-embedding-3-small (1536 dimensions, fast, cheap)
//   - text-embedding-3-large (3072 dimensions, high quality)
//   - text-embedding-ada-002 (1536 dimensions, legacy)
func NewOpenAIEmbedding(model, apiKey string) (*OpenAIEmbedding, error) {
	if model == "" {
		model = EmbeddingModelSmall // Default model
	}

	if apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required for embeddings\n\n" +
			"Fix:\n" +
			"  1. Set environment variable: export OPENAI_API_KEY=\"sk-...\"\n" +
			"  2. Or pass to constructor: NewOpenAIEmbedding(model, \"sk-...\")\n" +
			"  3. Get your key: https://platform.openai.com/api-keys")
	}

	client := openai.NewClient(option.WithAPIKey(apiKey))

	return &OpenAIEmbedding{
		client: &client,
		model:  model,
		config: DefaultEmbeddingConfig(),
	}, nil
}

// NewOpenAIEmbeddingWithClient creates a new OpenAI embedding provider with a custom client
func NewOpenAIEmbeddingWithClient(model string, client *openai.Client) *OpenAIEmbedding {
	if model == "" {
		model = EmbeddingModelSmall
	}

	return &OpenAIEmbedding{
		client: client,
		model:  model,
		config: DefaultEmbeddingConfig(),
	}
}

// WithConfig sets the embedding configuration
func (e *OpenAIEmbedding) WithConfig(config *EmbeddingConfig) *OpenAIEmbedding {
	e.config = config
	return e
}

// Embed generates an embedding for a single text
func (e *OpenAIEmbedding) Embed(ctx context.Context, text string) ([]float32, error) {
	if text == "" {
		return nil, fmt.Errorf("text cannot be empty\n\n" +
			"Fix: Provide non-empty text to embed")
	}

	// Preprocess text
	text = prepareTextForEmbedding(text, e.config)

	// Create embedding request
	response, err := e.client.Embeddings.New(ctx, openai.EmbeddingNewParams{
		Input: openai.EmbeddingNewParamsInputUnion{
			OfArrayOfStrings: []string{text},
		},
		Model: openai.EmbeddingModel(e.model),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w\n\n"+
			"Possible causes:\n"+
			"  - Network timeout (increase timeout with context deadline)\n"+
			"  - Rate limit exceeded (add retry logic or use WithDefaults())\n"+
			"  - Invalid API key (check OPENAI_API_KEY)\n"+
			"  - Text too long (max ~8,000 tokens for most models)\n", err)
	}

	if len(response.Data) == 0 {
		return nil, fmt.Errorf("no embedding returned from API")
	}

	// Convert float64 to float32
	embedding64 := response.Data[0].Embedding
	embedding := make([]float32, len(embedding64))
	for i, v := range embedding64 {
		embedding[i] = float32(v)
	}

	// Normalize if configured
	if e.config.Normalize {
		return NormalizeVector(embedding), nil
	}

	return embedding, nil
}

// EmbedBatch generates embeddings for multiple texts
func (e *OpenAIEmbedding) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("no texts provided")
	}

	// Preprocess all texts
	processedTexts := make([]string, len(texts))
	for i, text := range texts {
		processedTexts[i] = prepareTextForEmbedding(text, e.config)
	}

	// Process in batches if needed
	batchSize := e.config.BatchSize
	if batchSize <= 0 {
		batchSize = 100
	}

	allEmbeddings := make([][]float32, len(texts))

	for i := 0; i < len(processedTexts); i += batchSize {
		end := i + batchSize
		if end > len(processedTexts) {
			end = len(processedTexts)
		}

		batch := processedTexts[i:end]

		// Create embedding request for this batch
		response, err := e.client.Embeddings.New(ctx, openai.EmbeddingNewParams{
			Input: openai.EmbeddingNewParamsInputUnion{
				OfArrayOfStrings: batch,
			},
			Model: openai.EmbeddingModel(e.model),
		})

		if err != nil {
			return nil, fmt.Errorf("failed to generate embeddings for batch starting at %d: %w", i, err)
		}

		if len(response.Data) != len(batch) {
			return nil, fmt.Errorf("expected %d embeddings, got %d", len(batch), len(response.Data))
		}

		// Store embeddings (convert float64 to float32)
		for j, data := range response.Data {
			embedding64 := data.Embedding
			embedding := make([]float32, len(embedding64))
			for k, v := range embedding64 {
				embedding[k] = float32(v)
			}

			if e.config.Normalize {
				embedding = NormalizeVector(embedding)
			}
			allEmbeddings[i+j] = embedding
		}
	}

	return allEmbeddings, nil
}

// Dimensions returns the dimensionality of embeddings from this model
func (e *OpenAIEmbedding) Dimensions() int {
	switch e.model {
	case EmbeddingModelLarge:
		return 3072
	case EmbeddingModelSmall, EmbeddingModelAda002:
		return 1536
	default:
		// Default to 1536 for unknown models
		return 1536
	}
}

// Model returns the name of the embedding model
func (e *OpenAIEmbedding) Model() string {
	return e.model
}
