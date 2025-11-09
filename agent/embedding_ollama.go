package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	// Default Ollama embedding models
	OllamaEmbeddingModelNomic     = "nomic-embed-text"       // 768 dimensions, fast
	OllamaEmbeddingModelMxbai     = "mxbai-embed-large"      // 1024 dimensions
	OllamaEmbeddingModelAllMiniLM = "all-minilm"             // 384 dimensions
	DefaultOllamaURL              = "http://localhost:11434" // Default Ollama server
)

// OllamaEmbedding implements EmbeddingProvider using Ollama's embedding API
type OllamaEmbedding struct {
	baseURL string
	model   string
	config  *EmbeddingConfig
	client  *http.Client
}

// ollamaEmbeddingRequest represents the request to Ollama embedding API
type ollamaEmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

// ollamaEmbeddingResponse represents the response from Ollama embedding API
type ollamaEmbeddingResponse struct {
	Embedding []float64 `json:"embedding"`
}

// NewOllamaEmbedding creates a new Ollama embedding provider
// Supported models:
//   - nomic-embed-text (768 dimensions, recommended)
//   - mxbai-embed-large (1024 dimensions)
//   - all-minilm (384 dimensions, fast)
//
// baseURL: Ollama server URL (e.g., "http://localhost:11434")
func NewOllamaEmbedding(model, baseURL string) (*OllamaEmbedding, error) {
	if model == "" {
		model = OllamaEmbeddingModelNomic // Default model
	}

	if baseURL == "" {
		baseURL = DefaultOllamaURL
	}

	return &OllamaEmbedding{
		baseURL: baseURL,
		model:   model,
		config:  DefaultEmbeddingConfig(),
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}, nil
}

// WithConfig sets the embedding configuration
func (e *OllamaEmbedding) WithConfig(config *EmbeddingConfig) *OllamaEmbedding {
	e.config = config
	return e
}

// WithHTTPClient sets a custom HTTP client
func (e *OllamaEmbedding) WithHTTPClient(client *http.Client) *OllamaEmbedding {
	e.client = client
	return e
}

// Embed generates an embedding for a single text
func (e *OllamaEmbedding) Embed(ctx context.Context, text string) ([]float32, error) {
	if text == "" {
		return nil, fmt.Errorf("text cannot be empty")
	}

	// Preprocess text
	text = prepareTextForEmbedding(text, e.config)

	// Create request
	reqBody := ollamaEmbeddingRequest{
		Model:  e.model,
		Prompt: text,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Send request
	url := fmt.Sprintf("%s/api/embeddings", e.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Ollama API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var response ollamaEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Embedding) == 0 {
		return nil, fmt.Errorf("no embedding returned from API")
	}

	// Convert float64 to float32
	embedding := make([]float32, len(response.Embedding))
	for i, v := range response.Embedding {
		embedding[i] = float32(v)
	}

	// Normalize if configured
	if e.config.Normalize {
		return NormalizeVector(embedding), nil
	}

	return embedding, nil
}

// EmbedBatch generates embeddings for multiple texts
// Note: Ollama doesn't have a native batch API, so we process sequentially
func (e *OllamaEmbedding) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("no texts provided")
	}

	allEmbeddings := make([][]float32, len(texts))

	for i, text := range texts {
		embedding, err := e.Embed(ctx, text)
		if err != nil {
			return nil, fmt.Errorf("failed to embed text at index %d: %w", i, err)
		}
		allEmbeddings[i] = embedding
	}

	return allEmbeddings, nil
}

// Dimensions returns the dimensionality of embeddings from this model
func (e *OllamaEmbedding) Dimensions() int {
	switch e.model {
	case OllamaEmbeddingModelNomic:
		return 768
	case OllamaEmbeddingModelMxbai:
		return 1024
	case OllamaEmbeddingModelAllMiniLM:
		return 384
	default:
		// For unknown models, we need to get dimensions from an actual embedding
		// Return a reasonable default
		return 768
	}
}

// Model returns the name of the embedding model
func (e *OllamaEmbedding) Model() string {
	return e.model
}
