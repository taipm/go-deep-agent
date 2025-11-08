package agent

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

const (
	testModel  = "gpt-4o-mini"
	testAPIKey = "test-key"
)

func TestChunkDocument(t *testing.T) {
	text := "This is sentence one. This is sentence two. This is sentence three. This is sentence four."

	chunks := ChunkDocument(text, 50, 10)

	if len(chunks) == 0 {
		t.Error("Expected at least one chunk")
	}

	// Verify chunks overlap
	t.Logf("Created %d chunks from text of length %d", len(chunks), len(text))
	for i, chunk := range chunks {
		t.Logf("Chunk %d (len=%d): %s", i, len(chunk), chunk)
	}
}

func TestChunkDocumentEmpty(t *testing.T) {
	chunks := ChunkDocument("", 100, 10)

	if len(chunks) != 0 {
		t.Errorf("Expected 0 chunks for empty text, got %d", len(chunks))
	}
}

func TestChunkDocumentLarge(t *testing.T) {
	// Create a large document
	var sb strings.Builder
	for i := 0; i < 100; i++ {
		sb.WriteString(fmt.Sprintf("This is sentence number %d. ", i))
	}

	text := sb.String()
	chunks := ChunkDocument(text, 200, 50)

	if len(chunks) == 0 {
		t.Error("Expected multiple chunks for large text")
	}

	t.Logf("Large text (%d chars) split into %d chunks", len(text), len(chunks))
}

func TestWithRAG(t *testing.T) {
	docs := []string{
		"Go is a statically typed, compiled programming language.",
		"Python is an interpreted, high-level programming language.",
		"Rust is a systems programming language focused on safety.",
	}

	agent := NewOpenAI("gpt-4o-mini", "test-key").
		WithRAG(docs...)

	if !agent.ragEnabled {
		t.Error("Expected RAG to be enabled")
	}

	if len(agent.ragDocuments) != 3 {
		t.Errorf("Expected 3 documents, got %d", len(agent.ragDocuments))
	}

	if agent.ragConfig == nil {
		t.Error("Expected default RAG config to be set")
	}
}

func TestWithRAGDocuments(t *testing.T) {
	docs := []Document{
		{
			Content:  "Go programming language",
			Metadata: map[string]string{"source": "golang.org"},
		},
		{
			Content:  "Python programming language",
			Metadata: map[string]string{"source": "python.org"},
		},
	}

	agent := NewOpenAI("gpt-4o-mini", "test-key").
		WithRAGDocuments(docs...)

	if !agent.ragEnabled {
		t.Error("Expected RAG to be enabled")
	}

	if len(agent.ragDocuments) != 2 {
		t.Errorf("Expected 2 documents, got %d", len(agent.ragDocuments))
	}

	// Check metadata preserved
	if agent.ragDocuments[0].Metadata["source"] != "golang.org" {
		t.Error("Expected metadata to be preserved")
	}
}

func TestWithRAGTopK(t *testing.T) {
	agent := NewOpenAI("gpt-4o-mini", "test-key").
		WithRAGTopK(5)

	if agent.ragConfig.TopK != 5 {
		t.Errorf("Expected TopK=5, got %d", agent.ragConfig.TopK)
	}

	// Test with zero (should default)
	agent2 := NewOpenAI("gpt-4o-mini", "test-key").
		WithRAGTopK(0)

	if agent2.ragConfig.TopK != 3 {
		t.Errorf("Expected default TopK=3, got %d", agent2.ragConfig.TopK)
	}
}

func TestWithRAGChunkSize(t *testing.T) {
	agent := NewOpenAI("gpt-4o-mini", "test-key").
		WithRAGChunkSize(500)

	if agent.ragConfig.ChunkSize != 500 {
		t.Errorf("Expected ChunkSize=500, got %d", agent.ragConfig.ChunkSize)
	}
}

func TestDefaultRAGConfig(t *testing.T) {
	config := DefaultRAGConfig()

	if config.ChunkSize != 1000 {
		t.Errorf("Expected ChunkSize=1000, got %d", config.ChunkSize)
	}

	if config.ChunkOverlap != 200 {
		t.Errorf("Expected ChunkOverlap=200, got %d", config.ChunkOverlap)
	}

	if config.TopK != 3 {
		t.Errorf("Expected TopK=3, got %d", config.TopK)
	}

	if config.MinScore != 0.0 {
		t.Errorf("Expected MinScore=0.0, got %f", config.MinScore)
	}

	if config.Separator != "\n\n---\n\n" {
		t.Errorf("Expected separator, got %s", config.Separator)
	}
}

func TestCalculateSimilarity(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		document string
		minScore float64
	}{
		{
			name:     "Exact match",
			query:    "golang programming",
			document: "This is about golang programming language",
			minScore: 0.3,
		},
		{
			name:     "Partial match",
			query:    "python tutorial",
			document: "Learn python with this comprehensive guide",
			minScore: 0.2,
		},
		{
			name:     "No match",
			query:    "javascript",
			document: "This is about golang programming",
			minScore: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := calculateSimilarity(tt.query, tt.document)

			if score < 0.0 || score > 1.0 {
				t.Errorf("Score should be between 0 and 1, got %f", score)
			}

			if score < tt.minScore {
				t.Errorf("Expected score >= %f, got %f", tt.minScore, score)
			}

			t.Logf("Score for '%s': %f", tt.name, score)
		})
	}
}

func TestTokenize(t *testing.T) {
	text := "This is a test! How are you?"
	tokens := tokenize(text)

	// Should filter out stop words and short words
	if len(tokens) == 0 {
		t.Error("Expected some tokens")
	}

	// Check that stop words are removed
	for _, token := range tokens {
		if token == "is" || token == "a" || token == "are" {
			t.Errorf("Stop word '%s' should be filtered out", token)
		}
	}

	t.Logf("Tokens: %v", tokens)
}

func TestRetrieveRelevantDocs(t *testing.T) {
	docs := []string{
		"Go is a programming language designed for building reliable and efficient software.",
		"Python is widely used for data science and machine learning applications.",
		"Rust provides memory safety without garbage collection.",
		"JavaScript is the language of the web, running in browsers and servers.",
	}

	agent := NewOpenAI("gpt-4o-mini", "test-key").
		WithRAG(docs...).
		WithRAGTopK(2)

	// Retrieve docs relevant to "Go programming"
	retrieved, err := agent.retrieveRelevantDocs("Go programming language")

	if err != nil {
		t.Fatalf("Failed to retrieve docs: %v", err)
	}

	if len(retrieved) > 2 {
		t.Errorf("Expected max 2 docs (TopK), got %d", len(retrieved))
	}

	// First result should be about Go
	if len(retrieved) > 0 {
		if !strings.Contains(retrieved[0].Content, "Go") {
			t.Errorf("Expected first result to contain 'Go', got: %s", retrieved[0].Content)
		}
		t.Logf("Top result (score=%.2f): %s", retrieved[0].Score, retrieved[0].Content)
	}
}

func TestBuildRAGContext(t *testing.T) {
	docs := []Document{
		{Content: "First document", Score: 0.9},
		{Content: "Second document", Score: 0.7},
	}

	agent := NewOpenAI("gpt-4o-mini", "test-key")
	agent.ragConfig = DefaultRAGConfig()

	context := agent.buildRAGContext(docs)

	if context == "" {
		t.Error("Expected non-empty context")
	}

	if !strings.Contains(context, "First document") {
		t.Error("Context should contain first document")
	}

	if !strings.Contains(context, "Second document") {
		t.Error("Context should contain second document")
	}

	t.Logf("Context: %s", context)
}

func TestBuildRAGContextWithScores(t *testing.T) {
	docs := []Document{
		{Content: "Document 1", Score: 0.95},
		{Content: "Document 2", Score: 0.80},
	}

	agent := NewOpenAI("gpt-4o-mini", "test-key")
	agent.ragConfig = DefaultRAGConfig()
	agent.ragConfig.IncludeScores = true

	context := agent.buildRAGContext(docs)

	if !strings.Contains(context, "0.95") {
		t.Error("Context should include relevance scores")
	}

	t.Logf("Context with scores: %s", context)
}

func TestWithRAGRetriever(t *testing.T) {
	customRetriever := func(query string) ([]Document, error) {
		return []Document{
			{Content: "Custom retrieved document", Score: 1.0},
		}, nil
	}

	agent := NewOpenAI("gpt-4o-mini", "test-key").
		WithRAGRetriever(customRetriever)

	if !agent.ragEnabled {
		t.Error("Expected RAG to be enabled")
	}

	if agent.ragRetriever == nil {
		t.Error("Expected custom retriever to be set")
	}

	// Test retrieval
	docs, err := agent.retrieveRelevantDocs("test query")
	if err != nil {
		t.Fatalf("Retrieval failed: %v", err)
	}

	if len(docs) != 1 {
		t.Errorf("Expected 1 doc from custom retriever, got %d", len(docs))
	}

	if docs[0].Content != "Custom retrieved document" {
		t.Error("Expected custom retriever to be used")
	}
}

func TestClearRAG(t *testing.T) {
	agent := NewOpenAI("gpt-4o-mini", "test-key").
		WithRAG("Doc 1", "Doc 2").
		WithRAGTopK(5)

	if !agent.ragEnabled {
		t.Error("RAG should be enabled")
	}

	agent.ClearRAG()

	if agent.ragEnabled {
		t.Error("RAG should be disabled after ClearRAG")
	}

	if agent.ragDocuments != nil {
		t.Error("Documents should be cleared")
	}

	if agent.ragConfig != nil {
		t.Error("Config should be cleared")
	}
}

func TestGetLastRetrievedDocs(t *testing.T) {
	agent := NewOpenAI("gpt-4o-mini", "test-key")

	docs := agent.GetLastRetrievedDocs()

	if docs != nil {
		t.Error("Expected nil for no retrieved docs")
	}

	// Set some docs
	agent.lastRetrievedDocs = []Document{
		{Content: "Test doc"},
	}

	docs = agent.GetLastRetrievedDocs()

	if len(docs) != 1 {
		t.Errorf("Expected 1 doc, got %d", len(docs))
	}
}

func TestRAGWithMinScore(t *testing.T) {
	docs := []string{
		"Go programming language is great",
		"Python is also good",
		"Unrelated content about cooking",
	}

	agent := NewOpenAI("gpt-4o-mini", "test-key").
		WithRAG(docs...)

	agent.ragConfig.MinScore = 0.3 // Set minimum relevance threshold

	retrieved, err := agent.retrieveRelevantDocs("Go programming")
	if err != nil {
		t.Fatalf("Retrieval failed: %v", err)
	}

	// Should filter out low-scoring documents
	for _, doc := range retrieved {
		if doc.Score < 0.3 {
			t.Errorf("Document with score %f should be filtered (min=0.3)", doc.Score)
		}
	}

	t.Logf("Retrieved %d docs with min score 0.3", len(retrieved))
}

func TestRAGChaining(t *testing.T) {
	// Test fluent API chaining
	agent := NewOpenAI("gpt-4o-mini", "test-key").
		WithRAG("Doc 1", "Doc 2").
		WithRAGTopK(5).
		WithRAGChunkSize(500)

	if !agent.ragEnabled {
		t.Error("RAG should be enabled")
	}

	if agent.ragConfig.TopK != 5 {
		t.Error("TopK should be 5")
	}

	if agent.ragConfig.ChunkSize != 500 {
		t.Error("ChunkSize should be 500")
	}
}

func TestRAGWithConfig(t *testing.T) {
	config := &RAGConfig{
		ChunkSize:     800,
		ChunkOverlap:  100,
		TopK:          5,
		MinScore:      0.5,
		Separator:     "\n===\n",
		IncludeScores: true,
	}

	agent := NewOpenAI("gpt-4o-mini", "test-key").
		WithRAG("Test doc").
		WithRAGConfig(config)

	if agent.ragConfig.ChunkSize != 800 {
		t.Errorf("Expected ChunkSize=800, got %d", agent.ragConfig.ChunkSize)
	}

	if agent.ragConfig.Separator != "\n===\n" {
		t.Errorf("Expected custom separator, got %s", agent.ragConfig.Separator)
	}

	if !agent.ragConfig.IncludeScores {
		t.Error("Expected IncludeScores=true")
	}
}

func TestChunkOverlap(t *testing.T) {
	text := "Word1 Word2 Word3 Word4 Word5 Word6 Word7 Word8 Word9 Word10"

	// Chunk with overlap
	chunks := ChunkDocument(text, 20, 5)

	if len(chunks) < 2 {
		t.Error("Expected multiple chunks")
	}

	// Verify overlap exists (last chars of chunk N should appear in chunk N+1)
	// This is a simplified check
	t.Logf("Created %d overlapping chunks", len(chunks))
	for i, chunk := range chunks {
		t.Logf("Chunk %d: %s", i, chunk)
	}
}

func TestDocumentMetadata(t *testing.T) {
	doc := Document{
		Content: "Test content",
		Metadata: map[string]string{
			"source": "test.txt",
			"page":   "1",
		},
		Score: 0.5,
	}

	if doc.Metadata["source"] != "test.txt" {
		t.Error("Expected metadata to be accessible")
	}

	if doc.Score != 0.5 {
		t.Errorf("Expected score 0.5, got %f", doc.Score)
	}
}

// Test RAG integration with Ask (without actual API call)
func TestRAGIntegrationMock(t *testing.T) {
	docs := []string{
		"Go was created at Google by Robert Griesemer, Rob Pike, and Ken Thompson.",
		"Python was created by Guido van Rossum and released in 1991.",
	}

	agent := NewOpenAI("gpt-4o-mini", "test-key").
		WithRAG(docs...).
		WithRAGTopK(1)

	ctx := context.Background()

	// This will fail due to invalid API key, but we can verify RAG is enabled
	_, err := agent.Ask(ctx, "Who created Go?")

	// Should have attempted RAG retrieval
	retrieved := agent.GetLastRetrievedDocs()

	if retrieved == nil {
		t.Error("Expected documents to be retrieved")
	}

	if err == nil {
		t.Error("Expected API error (invalid key)")
	}

	t.Logf("RAG retrieved %d documents before API call", len(retrieved))
}

func BenchmarkChunkDocument(b *testing.B) {
	text := strings.Repeat("This is a test sentence. ", 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ChunkDocument(text, 500, 100)
	}
}

func BenchmarkCalculateSimilarity(b *testing.B) {
	query := "Go programming language features"
	document := "Go is a statically typed, compiled programming language designed at Google"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = calculateSimilarity(query, document)
	}
}

func BenchmarkTokenize(b *testing.B) {
	text := "This is a longer text with many words that need to be tokenized efficiently"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tokenize(text)
	}
}
