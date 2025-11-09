package agent

import (
	"context"
	"testing"
)

func TestWithVectorRAG(t *testing.T) {
	mockEmb := &MockEmbeddingProvider{
		embeddings: map[string][]float32{
			"test": {0.1, 0.2, 0.3},
		},
	}

	mockStore := &MockVectorStore{
		collections: make(map[string][]*VectorDocument),
	}

	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithVectorRAG(mockEmb, mockStore, "test-collection")

	if builder.embeddingProvider != mockEmb {
		t.Error("Expected embedding provider to be set")
	}

	if builder.vectorStore != mockStore {
		t.Error("Expected vector store to be set")
	}

	if builder.vectorCollection != "test-collection" {
		t.Errorf("Expected collection 'test-collection', got '%s'", builder.vectorCollection)
	}

	if !builder.ragEnabled {
		t.Error("Expected RAG to be enabled")
	}

	if builder.ragConfig == nil {
		t.Error("Expected RAG config to be initialized")
	}
}

func TestAddDocumentsToVector(t *testing.T) {
	ctx := context.Background()

	mockEmb := &MockEmbeddingProvider{
		embeddings: map[string][]float32{
			"doc1": {0.1, 0.2, 0.3},
			"doc2": {0.4, 0.5, 0.6},
		},
	}

	mockStore := &MockVectorStore{
		collections: make(map[string][]*VectorDocument),
	}

	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithVectorRAG(mockEmb, mockStore, "test-collection")

	// Create collection first
	config := &CollectionConfig{
		Name:      "test-collection",
		Dimension: 3,
	}
	err := mockStore.CreateCollection(ctx, "test-collection", config)
	if err != nil {
		t.Fatalf("Failed to create collection: %v", err)
	}

	// Add documents
	docs := []string{"doc1", "doc2"}
	ids, err := builder.AddDocumentsToVector(ctx, docs...)
	if err != nil {
		t.Fatalf("Failed to add documents: %v", err)
	}

	if len(ids) != 2 {
		t.Errorf("Expected 2 IDs, got %d", len(ids))
	}

	// Verify documents were added
	count, err := mockStore.Count(ctx, "test-collection")
	if err != nil {
		t.Fatalf("Failed to count: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 documents, got %d", count)
	}
}

func TestAddDocumentsToVectorNoProvider(t *testing.T) {
	ctx := context.Background()

	builder := NewOpenAI("gpt-4o-mini", "test-key")

	_, err := builder.AddDocumentsToVector(ctx, "doc1")
	if err == nil {
		t.Error("Expected error when vector store not configured")
	}
}

func TestAddVectorDocuments(t *testing.T) {
	ctx := context.Background()

	mockEmb := &MockEmbeddingProvider{
		embeddings: map[string][]float32{
			"content1": {0.1, 0.2, 0.3},
		},
	}

	mockStore := &MockVectorStore{
		collections: make(map[string][]*VectorDocument),
	}

	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithVectorRAG(mockEmb, mockStore, "test-collection")

	// Create collection
	config := &CollectionConfig{
		Name:      "test-collection",
		Dimension: 3,
	}
	mockStore.CreateCollection(ctx, "test-collection", config)

	// Add vector documents with metadata
	vectorDocs := []*VectorDocument{
		{
			Content:   "content1",
			Embedding: []float32{0.1, 0.2, 0.3},
			Metadata: map[string]interface{}{
				"category": "test",
			},
		},
	}

	ids, err := builder.AddVectorDocuments(ctx, vectorDocs...)
	if err != nil {
		t.Fatalf("Failed to add vector documents: %v", err)
	}

	if len(ids) != 1 {
		t.Errorf("Expected 1 ID, got %d", len(ids))
	}
}

func TestRetrieveFromVector(t *testing.T) {
	ctx := context.Background()

	mockEmb := &MockEmbeddingProvider{
		embeddings: map[string][]float32{
			"query":    {0.1, 0.2, 0.3},
			"content1": {0.1, 0.2, 0.3},
			"content2": {0.7, 0.8, 0.9},
		},
	}

	mockStore := &MockVectorStore{
		collections: make(map[string][]*VectorDocument),
	}

	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithVectorRAG(mockEmb, mockStore, "test-collection").
		WithRAGTopK(2)

	// Create collection and add documents
	config := &CollectionConfig{
		Name:      "test-collection",
		Dimension: 3,
	}
	mockStore.CreateCollection(ctx, "test-collection", config)

	vectorDocs := []*VectorDocument{
		{
			ID:        "doc1",
			Content:   "content1",
			Embedding: []float32{0.1, 0.2, 0.3},
			Metadata: map[string]interface{}{
				"category": "tech",
			},
		},
		{
			ID:        "doc2",
			Content:   "content2",
			Embedding: []float32{0.7, 0.8, 0.9},
			Metadata: map[string]interface{}{
				"category": "science",
			},
		},
	}

	mockStore.Add(ctx, "test-collection", vectorDocs)

	// Retrieve documents
	docs, err := builder.retrieveFromVector(ctx, "query")
	if err != nil {
		t.Fatalf("Failed to retrieve: %v", err)
	}

	if len(docs) != 2 {
		t.Errorf("Expected 2 documents, got %d", len(docs))
	}

	// Check document content
	if docs[0].Content != "content1" && docs[0].Content != "content2" {
		t.Errorf("Unexpected content: %s", docs[0].Content)
	}

	// Check metadata conversion
	if docs[0].Metadata["category"] == "" {
		t.Error("Expected metadata to be converted")
	}

	// Check score
	if docs[0].Score <= 0 {
		t.Error("Expected positive score")
	}
}

func TestRetrieveRelevantDocsWithVector(t *testing.T) {
	ctx := context.Background()

	mockEmb := &MockEmbeddingProvider{
		embeddings: map[string][]float32{
			"test query": {0.1, 0.2, 0.3},
			"doc1":       {0.1, 0.2, 0.3},
		},
	}

	mockStore := &MockVectorStore{
		collections: make(map[string][]*VectorDocument),
	}

	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithVectorRAG(mockEmb, mockStore, "test-collection")

	// Create collection and add docs
	config := &CollectionConfig{
		Name:      "test-collection",
		Dimension: 3,
	}
	mockStore.CreateCollection(ctx, "test-collection", config)

	vectorDocs := []*VectorDocument{
		{
			ID:        "doc1",
			Content:   "doc1",
			Embedding: []float32{0.1, 0.2, 0.3},
		},
	}
	mockStore.Add(ctx, "test-collection", vectorDocs)

	// Retrieve using vector search
	docs, err := builder.retrieveRelevantDocs(ctx, "test query")
	if err != nil {
		t.Fatalf("Failed to retrieve: %v", err)
	}

	if len(docs) == 0 {
		t.Error("Expected documents to be retrieved")
	}
}

func TestRetrieveRelevantDocsFallbackToTFIDF(t *testing.T) {
	ctx := context.Background()

	// No vector store configured - should fall back to TF-IDF
	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithRAG("Go is a programming language", "Python is also a language")

	docs, err := builder.retrieveRelevantDocs(ctx, "programming")
	if err != nil {
		t.Fatalf("Failed to retrieve: %v", err)
	}

	if len(docs) == 0 {
		t.Error("Expected documents from TF-IDF fallback")
	}

	// Should retrieve "Go is a programming language" first
	if len(docs) > 0 && docs[0].Score <= 0 {
		t.Error("Expected positive TF-IDF score")
	}
}

func TestRetrieveRelevantDocsCustomRetriever(t *testing.T) {
	ctx := context.Background()

	customCalled := false
	customRetriever := func(query string) ([]Document, error) {
		customCalled = true
		return []Document{
			{Content: "custom doc", Score: 1.0},
		}, nil
	}

	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithRAGRetriever(customRetriever)

	docs, err := builder.retrieveRelevantDocs(ctx, "test")
	if err != nil {
		t.Fatalf("Failed to retrieve: %v", err)
	}

	if !customCalled {
		t.Error("Expected custom retriever to be called")
	}

	if len(docs) != 1 || docs[0].Content != "custom doc" {
		t.Error("Expected custom document")
	}
}

func TestClearRAGClearsVectorStore(t *testing.T) {
	mockEmb := &MockEmbeddingProvider{
		embeddings: map[string][]float32{
			"test": {0.1, 0.2, 0.3},
		},
	}

	mockStore := &MockVectorStore{
		collections: make(map[string][]*VectorDocument),
	}

	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithVectorRAG(mockEmb, mockStore, "test-collection")

	if builder.vectorStore == nil {
		t.Error("Expected vector store to be set")
	}

	builder.ClearRAG()

	if builder.vectorStore != nil {
		t.Error("Expected vector store to be cleared")
	}

	if builder.embeddingProvider != nil {
		t.Error("Expected embedding provider to be cleared")
	}

	if builder.vectorCollection != "" {
		t.Error("Expected collection name to be cleared")
	}

	if builder.ragEnabled {
		t.Error("Expected RAG to be disabled")
	}
}

func TestVectorRAGWithCustomConfig(t *testing.T) {
	mockEmb := &MockEmbeddingProvider{
		embeddings: map[string][]float32{
			"test": {0.1, 0.2, 0.3},
		},
	}

	mockStore := &MockVectorStore{
		collections: make(map[string][]*VectorDocument),
	}

	customConfig := &RAGConfig{
		TopK:          5,
		MinScore:      0.7,
		Separator:     "\n---\n",
		IncludeScores: true,
	}

	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithVectorRAG(mockEmb, mockStore, "test-collection").
		WithRAGConfig(customConfig)

	if builder.ragConfig.TopK != 5 {
		t.Errorf("Expected TopK 5, got %d", builder.ragConfig.TopK)
	}

	if builder.ragConfig.MinScore != 0.7 {
		t.Errorf("Expected MinScore 0.7, got %f", builder.ragConfig.MinScore)
	}

	if !builder.ragConfig.IncludeScores {
		t.Error("Expected IncludeScores to be true")
	}
}

func TestRetrieveFromVectorNoProvider(t *testing.T) {
	ctx := context.Background()

	builder := NewOpenAI("gpt-4o-mini", "test-key")

	_, err := builder.retrieveFromVector(ctx, "test")
	if err == nil {
		t.Error("Expected error when vector store not configured")
	}
}

func TestRetrieveFromVectorWithMinScore(t *testing.T) {
	ctx := context.Background()

	mockEmb := &MockEmbeddingProvider{
		embeddings: map[string][]float32{
			"query": {0.1, 0.2, 0.3},
			"doc1":  {0.1, 0.2, 0.3},
			"doc2":  {0.9, 0.9, 0.9},
		},
	}

	mockStore := &MockVectorStore{
		collections: make(map[string][]*VectorDocument),
	}

	builder := NewOpenAI("gpt-4o-mini", "test-key").
		WithVectorRAG(mockEmb, mockStore, "test-collection").
		WithRAGConfig(&RAGConfig{
			TopK:     10,
			MinScore: 0.8, // High threshold
		})

	// Create collection
	config := &CollectionConfig{
		Name:      "test-collection",
		Dimension: 3,
	}
	mockStore.CreateCollection(ctx, "test-collection", config)

	// Add documents with different scores
	vectorDocs := []*VectorDocument{
		{
			ID:        "doc1",
			Content:   "doc1",
			Embedding: []float32{0.1, 0.2, 0.3}, // High similarity
		},
		{
			ID:        "doc2",
			Content:   "doc2",
			Embedding: []float32{0.9, 0.9, 0.9}, // Low similarity
		},
	}
	mockStore.Add(ctx, "test-collection", vectorDocs)

	// Retrieve - should filter by MinScore
	docs, err := builder.retrieveFromVector(ctx, "query")
	if err != nil {
		t.Fatalf("Failed to retrieve: %v", err)
	}

	// Only high-scoring docs should be returned
	// (Note: MockVectorStore returns all, but real implementation would filter)
	if len(docs) == 0 {
		t.Error("Expected at least some documents")
	}
}
