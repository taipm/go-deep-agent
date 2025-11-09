package agent

import (
	"context"
	"math"
	"strings"
	"testing"
)

// TestCosineSimilarity tests the cosine similarity calculation
func TestCosineSimilarity(t *testing.T) {
	tests := []struct {
		name      string
		a         []float32
		b         []float32
		expected  float32
		tolerance float32
	}{
		{
			name:      "identical vectors",
			a:         []float32{1, 0, 0},
			b:         []float32{1, 0, 0},
			expected:  1.0,
			tolerance: 0.0001,
		},
		{
			name:      "opposite vectors",
			a:         []float32{1, 0, 0},
			b:         []float32{-1, 0, 0},
			expected:  -1.0,
			tolerance: 0.0001,
		},
		{
			name:      "orthogonal vectors",
			a:         []float32{1, 0, 0},
			b:         []float32{0, 1, 0},
			expected:  0.0,
			tolerance: 0.0001,
		},
		{
			name:      "45 degree angle",
			a:         []float32{1, 1, 0},
			b:         []float32{1, 0, 0},
			expected:  0.7071,
			tolerance: 0.001,
		},
		{
			name:      "multi-dimensional similar",
			a:         []float32{1, 2, 3, 4},
			b:         []float32{2, 4, 6, 8},
			expected:  1.0,
			tolerance: 0.0001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CosineSimilarity(tt.a, tt.b)
			if err != nil {
				t.Fatalf("CosineSimilarity() error = %v", err)
			}
			if math.Abs(float64(result-tt.expected)) > float64(tt.tolerance) {
				t.Errorf("CosineSimilarity(%v, %v) = %f, want %f (tolerance %f)",
					tt.a, tt.b, result, tt.expected, tt.tolerance)
			}
		})
	}
}

// TestCosineSimilarityEdgeCases tests edge cases
func TestCosineSimilarityEdgeCases(t *testing.T) {
	tests := []struct {
		name string
		a    []float32
		b    []float32
	}{
		{
			name: "empty vectors",
			a:    []float32{},
			b:    []float32{},
		},
		{
			name: "zero vectors",
			a:    []float32{0, 0, 0},
			b:    []float32{0, 0, 0},
		},
		{
			name: "mismatched dimensions",
			a:    []float32{1, 2},
			b:    []float32{1, 2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CosineSimilarity(tt.a, tt.b)
			// These should return errors
			if err == nil {
				t.Errorf("CosineSimilarity(%v, %v) should return error, got result %f",
					tt.a, tt.b, result)
			}
		})
	}
}

// TestDotProduct tests the dot product calculation
func TestDotProduct(t *testing.T) {
	tests := []struct {
		name     string
		a        []float32
		b        []float32
		expected float32
	}{
		{
			name:     "simple vectors",
			a:        []float32{1, 2, 3},
			b:        []float32{4, 5, 6},
			expected: 32, // 1*4 + 2*5 + 3*6 = 4 + 10 + 18 = 32
		},
		{
			name:     "negative values",
			a:        []float32{1, -2, 3},
			b:        []float32{-1, 2, -3},
			expected: -14, // 1*(-1) + (-2)*2 + 3*(-3) = -1 - 4 - 9 = -14
		},
		{
			name:     "zero result",
			a:        []float32{1, 0, -1},
			b:        []float32{1, 5, 1},
			expected: 0, // 1*1 + 0*5 + (-1)*1 = 1 + 0 - 1 = 0
		},
		{
			name:     "empty vectors",
			a:        []float32{},
			b:        []float32{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := DotProduct(tt.a, tt.b)
			if err != nil && tt.expected != 0 {
				t.Fatalf("DotProduct() error = %v", err)
			}
			if result != tt.expected {
				t.Errorf("DotProduct(%v, %v) = %f, want %f", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// TestEuclideanDistance tests the Euclidean distance calculation
func TestEuclideanDistance(t *testing.T) {
	tests := []struct {
		name      string
		a         []float32
		b         []float32
		expected  float32
		tolerance float32
	}{
		{
			name:      "identical vectors",
			a:         []float32{1, 2, 3},
			b:         []float32{1, 2, 3},
			expected:  0.0,
			tolerance: 0.0001,
		},
		{
			name:      "simple distance",
			a:         []float32{0, 0, 0},
			b:         []float32{3, 4, 0},
			expected:  5.0, // sqrt(9 + 16) = 5
			tolerance: 0.0001,
		},
		{
			name:      "negative coordinates",
			a:         []float32{-1, -1, -1},
			b:         []float32{1, 1, 1},
			expected:  3.464, // sqrt(4 + 4 + 4) ≈ 3.464
			tolerance: 0.01,
		},
		{
			name:      "1D distance",
			a:         []float32{0},
			b:         []float32{5},
			expected:  5.0,
			tolerance: 0.0001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := EuclideanDistance(tt.a, tt.b)
			if err != nil {
				t.Fatalf("EuclideanDistance() error = %v", err)
			}
			if math.Abs(float64(result-tt.expected)) > float64(tt.tolerance) {
				t.Errorf("EuclideanDistance(%v, %v) = %f, want %f (tolerance %f)",
					tt.a, tt.b, result, tt.expected, tt.tolerance)
			}
		})
	}
}

// TestNormalizeVector tests vector normalization
func TestNormalizeVector(t *testing.T) {
	tests := []struct {
		name      string
		input     []float32
		tolerance float32
	}{
		{
			name:      "simple vector",
			input:     []float32{3, 4},
			tolerance: 0.0001,
		},
		{
			name:      "already normalized",
			input:     []float32{1, 0, 0},
			tolerance: 0.0001,
		},
		{
			name:      "3D vector",
			input:     []float32{1, 2, 3},
			tolerance: 0.0001,
		},
		{
			name:      "negative values",
			input:     []float32{-1, -2, -3},
			tolerance: 0.0001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeVector(tt.input)

			// Check that the magnitude is 1
			magnitude := float32(0)
			for _, v := range result {
				magnitude += v * v
			}
			magnitude = sqrt32(magnitude)

			if math.Abs(float64(magnitude-1.0)) > float64(tt.tolerance) {
				t.Errorf("NormalizeVector(%v) magnitude = %f, want 1.0", tt.input, magnitude)
			}

			// Check that direction is preserved (by checking ratio)
			if len(result) > 0 && tt.input[0] != 0 {
				ratio := result[0] / tt.input[0]
				for i := 1; i < len(result); i++ {
					if tt.input[i] != 0 {
						currentRatio := result[i] / tt.input[i]
						if math.Abs(float64(ratio-currentRatio)) > float64(tt.tolerance) {
							t.Errorf("NormalizeVector(%v) changed direction", tt.input)
							break
						}
					}
				}
			}
		})
	}
}

// TestNormalizeVectorZero tests normalization of zero vector
func TestNormalizeVectorZero(t *testing.T) {
	input := []float32{0, 0, 0}
	result := NormalizeVector(input)

	for _, v := range result {
		if !math.IsNaN(float64(v)) && v != 0 {
			t.Errorf("NormalizeVector(zero vector) should return zeros or NaN, got %v", result)
		}
	}
}

// TestSqrt32 tests the fast square root function
func TestSqrt32(t *testing.T) {
	tests := []struct {
		name      string
		input     float32
		tolerance float32
	}{
		{
			name:      "perfect square",
			input:     16.0,
			tolerance: 0.001,
		},
		{
			name:      "small number",
			input:     2.0,
			tolerance: 0.001,
		},
		{
			name:      "large number",
			input:     1000.0,
			tolerance: 0.01,
		},
		{
			name:      "fractional",
			input:     0.25,
			tolerance: 0.001,
		},
		{
			name:      "zero",
			input:     0.0,
			tolerance: 0.001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sqrt32(tt.input)
			expected := float32(math.Sqrt(float64(tt.input)))

			if math.Abs(float64(result-expected)) > float64(tt.tolerance) {
				t.Errorf("sqrt32(%f) = %f, want %f (tolerance %f)",
					tt.input, result, expected, tt.tolerance)
			}
		})
	}
}

// TestPrepareTextForEmbedding tests text preprocessing
func TestPrepareTextForEmbedding(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		config   *EmbeddingConfig
		expected string
	}{
		{
			name:     "no processing",
			input:    "Hello World",
			config:   &EmbeddingConfig{StripNewlines: false},
			expected: "Hello World",
		},
		{
			name:     "strip newlines",
			input:    "Hello\nWorld\nTest",
			config:   &EmbeddingConfig{StripNewlines: true},
			expected: "Hello World Test",
		},
		{
			name:     "multiple newlines",
			input:    "Hello\n\n\nWorld",
			config:   &EmbeddingConfig{StripNewlines: true},
			expected: "Hello World",
		},
		{
			name:     "mixed whitespace",
			input:    "Hello\n\tWorld  Test",
			config:   &EmbeddingConfig{StripNewlines: true},
			expected: "Hello World Test",
		},
		{
			name:     "trim spaces",
			input:    "  Hello World  ",
			config:   &EmbeddingConfig{StripNewlines: false},
			expected: "Hello World",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := prepareTextForEmbedding(tt.input, tt.config)
			if result != tt.expected {
				t.Errorf("prepareTextForEmbedding(%q) = %q, want %q",
					tt.input, result, tt.expected)
			}
		})
	}
}

// TestDefaultEmbeddingConfig tests the default configuration
func TestDefaultEmbeddingConfig(t *testing.T) {
	config := DefaultEmbeddingConfig()

	if config.BatchSize != 100 {
		t.Errorf("Default BatchSize = %d, want 100", config.BatchSize)
	}

	if config.Normalize {
		t.Error("Default Normalize should be false")
	}

	if !config.StripNewlines {
		t.Error("Default StripNewlines should be true")
	}
}

// MockEmbeddingProvider is a mock implementation for testing
type MockEmbeddingProvider struct {
	embeddings map[string][]float32
	model      string
	dimensions int
}

func NewMockEmbeddingProvider(model string, dimensions int) *MockEmbeddingProvider {
	return &MockEmbeddingProvider{
		embeddings: make(map[string][]float32),
		model:      model,
		dimensions: dimensions,
	}
}

func (m *MockEmbeddingProvider) AddEmbedding(text string, embedding []float32) {
	m.embeddings[text] = embedding
}

func (m *MockEmbeddingProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	if embedding, ok := m.embeddings[text]; ok {
		return embedding, nil
	}
	// Return a simple deterministic embedding based on text length
	embedding := make([]float32, m.dimensions)
	for i := range embedding {
		embedding[i] = float32(len(text)) / float32(m.dimensions)
	}
	return embedding, nil
}

func (m *MockEmbeddingProvider) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	results := make([][]float32, len(texts))
	for i, text := range texts {
		embedding, err := m.Embed(ctx, text)
		if err != nil {
			return nil, err
		}
		results[i] = embedding
	}
	return results, nil
}

func (m *MockEmbeddingProvider) Dimensions() int {
	return m.dimensions
}

func (m *MockEmbeddingProvider) Model() string {
	return m.model
}

// TestMockEmbeddingProvider tests the mock provider
func TestMockEmbeddingProvider(t *testing.T) {
	ctx := context.Background()
	provider := NewMockEmbeddingProvider("mock-model", 128)

	// Test model and dimensions
	if provider.Model() != "mock-model" {
		t.Errorf("Model() = %s, want mock-model", provider.Model())
	}

	if provider.Dimensions() != 128 {
		t.Errorf("Dimensions() = %d, want 128", provider.Dimensions())
	}

	// Test single embedding
	text := "test text"
	embedding, err := provider.Embed(ctx, text)
	if err != nil {
		t.Fatalf("Embed() error = %v", err)
	}

	if len(embedding) != 128 {
		t.Errorf("Embedding length = %d, want 128", len(embedding))
	}

	// Test batch embedding
	texts := []string{"text1", "text2", "text3"}
	embeddings, err := provider.EmbedBatch(ctx, texts)
	if err != nil {
		t.Fatalf("EmbedBatch() error = %v", err)
	}

	if len(embeddings) != 3 {
		t.Errorf("EmbedBatch() returned %d embeddings, want 3", len(embeddings))
	}

	for i, emb := range embeddings {
		if len(emb) != 128 {
			t.Errorf("Embedding[%d] length = %d, want 128", i, len(emb))
		}
	}
}

// TestEmbeddingProviderInterface tests that implementations satisfy the interface
func TestEmbeddingProviderInterface(t *testing.T) {
	var _ EmbeddingProvider = (*MockEmbeddingProvider)(nil)
}

// BenchmarkCosineSimilarity benchmarks the cosine similarity calculation
func BenchmarkCosineSimilarity(b *testing.B) {
	a := make([]float32, 1536)
	vec := make([]float32, 1536)
	for i := range a {
		a[i] = float32(i)
		vec[i] = float32(i) * 2
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CosineSimilarity(a, vec)
	}
}

// BenchmarkNormalizeVector benchmarks vector normalization
func BenchmarkNormalizeVector(b *testing.B) {
	v := make([]float32, 1536)
	for i := range v {
		v[i] = float32(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NormalizeVector(v)
	}
}

// BenchmarkSqrt32 benchmarks the fast square root
func BenchmarkSqrt32(b *testing.B) {
	values := []float32{1.0, 2.0, 10.0, 100.0, 1000.0}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, v := range values {
			sqrt32(v)
		}
	}
}

// TestEmbeddingIntegration tests basic integration scenarios
func TestEmbeddingIntegration(t *testing.T) {
	ctx := context.Background()
	provider := NewMockEmbeddingProvider("test-model", 384)

	// Add some predefined embeddings
	provider.AddEmbedding("hello", []float32{1, 0, 0})
	provider.AddEmbedding("world", []float32{0, 1, 0})
	provider.AddEmbedding("test", []float32{0, 0, 1})

	// Test similarity between known embeddings
	emb1, _ := provider.Embed(ctx, "hello")
	emb2, _ := provider.Embed(ctx, "world")
	emb3, _ := provider.Embed(ctx, "hello")

	// Same text should have similarity 1.0
	sim, err := CosineSimilarity(emb1, emb3)
	if err != nil {
		t.Fatalf("CosineSimilarity() error = %v", err)
	}
	if math.Abs(float64(sim-1.0)) > 0.0001 {
		t.Errorf("Same text similarity = %f, want 1.0", sim)
	}

	// Different texts should have lower similarity
	sim, err = CosineSimilarity(emb1, emb2)
	if err != nil {
		t.Fatalf("CosineSimilarity() error = %v", err)
	}
	if sim >= 0.9 {
		t.Errorf("Different text similarity = %f, should be < 0.9", sim)
	}
}

// TestEmbeddingConfigNormalization tests normalization behavior
func TestEmbeddingConfigNormalization(t *testing.T) {
	vector := []float32{3, 4} // Magnitude is 5

	// Test without normalization
	config := &EmbeddingConfig{Normalize: false}
	_ = config // Use config in actual embedding call

	// Test with normalization
	normalized := NormalizeVector(vector)
	expectedX := float32(3.0 / 5.0) // 0.6
	expectedY := float32(4.0 / 5.0) // 0.8

	if math.Abs(float64(normalized[0]-expectedX)) > 0.0001 {
		t.Errorf("Normalized X = %f, want %f", normalized[0], expectedX)
	}

	if math.Abs(float64(normalized[1]-expectedY)) > 0.0001 {
		t.Errorf("Normalized Y = %f, want %f", normalized[1], expectedY)
	}
}

// TestTextPreparationWithNewlines tests text preprocessing with newlines
func TestTextPreparationWithNewlines(t *testing.T) {
	config := &EmbeddingConfig{StripNewlines: true}

	input := `This is a multiline
text with various
newline characters`

	result := prepareTextForEmbedding(input, config)

	if strings.Contains(result, "\n") {
		t.Errorf("Result still contains newlines: %q", result)
	}

	// Should have spaces instead
	if !strings.Contains(result, " ") {
		t.Error("Result should contain spaces between words")
	}
}
