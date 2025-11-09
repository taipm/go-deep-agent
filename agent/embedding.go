package agent

import (
	"context"
	"fmt"
	"strings"
)

// EmbeddingProvider is the interface for generating text embeddings
type EmbeddingProvider interface {
	// Embed generates an embedding vector for a single text
	Embed(ctx context.Context, text string) ([]float32, error)

	// EmbedBatch generates embedding vectors for multiple texts
	EmbedBatch(ctx context.Context, texts []string) ([][]float32, error)

	// Dimensions returns the dimensionality of the embedding vectors
	Dimensions() int

	// Model returns the name of the embedding model
	Model() string
}

// EmbeddingResult represents the result of an embedding operation
type EmbeddingResult struct {
	Text      string    // Original text
	Embedding []float32 // Generated embedding vector
	Index     int       // Original index in batch
	Error     error     // Error if embedding failed
}

// EmbeddingConfig holds configuration for embedding operations
type EmbeddingConfig struct {
	// BatchSize is the maximum number of texts to embed in a single request
	BatchSize int

	// Normalize determines whether to normalize embeddings to unit length
	Normalize bool

	// StripNewlines removes newlines from text before embedding
	StripNewlines bool
}

// DefaultEmbeddingConfig returns the default embedding configuration
func DefaultEmbeddingConfig() *EmbeddingConfig {
	return &EmbeddingConfig{
		BatchSize:     100,   // Process up to 100 texts per batch
		Normalize:     false, // Don't normalize by default (model-dependent)
		StripNewlines: true,  // Strip newlines by default
	}
}

// CosineSimilarity calculates the cosine similarity between two vectors
// Returns a value between -1 and 1, where 1 means identical direction
func CosineSimilarity(a, b []float32) (float32, error) {
	if len(a) != len(b) {
		return 0, fmt.Errorf("vectors must have same dimensions: %d vs %d", len(a), len(b))
	}

	var dotProduct, normA, normB float32

	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0, fmt.Errorf("cannot compute similarity with zero vector")
	}

	// Calculate cosine similarity
	similarity := dotProduct / (sqrt32(normA) * sqrt32(normB))

	return similarity, nil
}

// DotProduct calculates the dot product of two vectors
func DotProduct(a, b []float32) (float32, error) {
	if len(a) != len(b) {
		return 0, fmt.Errorf("vectors must have same dimensions: %d vs %d", len(a), len(b))
	}

	var product float32
	for i := 0; i < len(a); i++ {
		product += a[i] * b[i]
	}

	return product, nil
}

// EuclideanDistance calculates the Euclidean distance between two vectors
func EuclideanDistance(a, b []float32) (float32, error) {
	if len(a) != len(b) {
		return 0, fmt.Errorf("vectors must have same dimensions: %d vs %d", len(a), len(b))
	}

	var sum float32
	for i := 0; i < len(a); i++ {
		diff := a[i] - b[i]
		sum += diff * diff
	}

	return sqrt32(sum), nil
}

// NormalizeVector normalizes a vector to unit length
func NormalizeVector(v []float32) []float32 {
	var norm float32
	for _, val := range v {
		norm += val * val
	}

	norm = sqrt32(norm)
	if norm == 0 {
		return v
	}

	normalized := make([]float32, len(v))
	for i, val := range v {
		normalized[i] = val / norm
	}

	return normalized
}

// sqrt32 is a helper function for float32 square root
func sqrt32(x float32) float32 {
	// Simple Newton-Raphson method for square root
	if x == 0 {
		return 0
	}

	// Initial guess
	z := x
	for i := 0; i < 10; i++ {
		z = (z + x/z) / 2
	}

	return z
}

// prepareTextForEmbedding preprocesses text according to config
func prepareTextForEmbedding(text string, config *EmbeddingConfig) string {
	if config == nil {
		config = DefaultEmbeddingConfig()
	}

	// Trim leading and trailing whitespace
	text = strings.TrimSpace(text)

	if config.StripNewlines {
		// Replace newlines and tabs with spaces
		text = strings.ReplaceAll(text, "\n", " ")
		text = strings.ReplaceAll(text, "\r", " ")
		text = strings.ReplaceAll(text, "\t", " ")

		// Collapse multiple spaces into single space
		for strings.Contains(text, "  ") {
			text = strings.ReplaceAll(text, "  ", " ")
		}
	}

	return text
}
