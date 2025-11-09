package agent

import (
	"context"
	"testing"
	"time"
)

// TestVectorDocumentCreation tests creating vector documents
func TestVectorDocumentCreation(t *testing.T) {
	now := time.Now()
	doc := &VectorDocument{
		ID:        "test-1",
		Content:   "This is a test document",
		Embedding: []float32{0.1, 0.2, 0.3},
		Metadata: map[string]interface{}{
			"source": "test",
			"author": "tester",
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	if doc.ID != "test-1" {
		t.Errorf("Expected ID 'test-1', got %s", doc.ID)
	}

	if doc.Content != "This is a test document" {
		t.Errorf("Unexpected content: %s", doc.Content)
	}

	if len(doc.Embedding) != 3 {
		t.Errorf("Expected 3-dimensional embedding, got %d", len(doc.Embedding))
	}

	if doc.Metadata["source"] != "test" {
		t.Errorf("Expected metadata source 'test', got %v", doc.Metadata["source"])
	}
}

// TestCollectionConfig tests collection configuration
func TestCollectionConfig(t *testing.T) {
	config := &CollectionConfig{
		Name:           "test-collection",
		Description:    "Test collection for unit tests",
		Dimension:      384,
		DistanceMetric: DistanceMetricCosine,
	}

	if config.Name != "test-collection" {
		t.Errorf("Expected name 'test-collection', got %s", config.Name)
	}

	if config.Dimension != 384 {
		t.Errorf("Expected dimension 384, got %d", config.Dimension)
	}

	if config.DistanceMetric != DistanceMetricCosine {
		t.Errorf("Expected cosine metric, got %s", config.DistanceMetric)
	}
}

// TestDistanceMetrics tests distance metric constants
func TestDistanceMetrics(t *testing.T) {
	tests := []struct {
		name   string
		metric DistanceMetric
		want   string
	}{
		{"cosine", DistanceMetricCosine, "cosine"},
		{"euclidean", DistanceMetricEuclidean, "euclidean"},
		{"dot_product", DistanceMetricDotProduct, "dot_product"},
		{"l2", DistanceMetricL2, "l2"},
		{"ip", DistanceMetricIP, "ip"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.metric) != tt.want {
				t.Errorf("Expected %s, got %s", tt.want, string(tt.metric))
			}
		})
	}
}

// TestSearchRequestDefaults tests default search request creation
func TestSearchRequestDefaults(t *testing.T) {
	queryVec := []float32{0.1, 0.2, 0.3}
	req := DefaultSearchRequest("test-collection", queryVec)

	if req.Collection != "test-collection" {
		t.Errorf("Expected collection 'test-collection', got %s", req.Collection)
	}

	if req.TopK != 10 {
		t.Errorf("Expected TopK 10, got %d", req.TopK)
	}

	if !req.IncludeMetadata {
		t.Error("Expected IncludeMetadata to be true")
	}

	if !req.IncludeContent {
		t.Error("Expected IncludeContent to be true")
	}

	if req.IncludeEmbedding {
		t.Error("Expected IncludeEmbedding to be false")
	}

	if req.MinScore != 0.0 {
		t.Errorf("Expected MinScore 0.0, got %f", req.MinScore)
	}

	if len(req.QueryVector) != 3 {
		t.Errorf("Expected query vector length 3, got %d", len(req.QueryVector))
	}
}

// TestTextSearchRequestDefaults tests default text search request creation
func TestTextSearchRequestDefaults(t *testing.T) {
	req := DefaultTextSearchRequest("test-collection", "test query")

	if req.Collection != "test-collection" {
		t.Errorf("Expected collection 'test-collection', got %s", req.Collection)
	}

	if req.Query != "test query" {
		t.Errorf("Expected query 'test query', got %s", req.Query)
	}

	if req.TopK != 10 {
		t.Errorf("Expected TopK 10, got %d", req.TopK)
	}

	if !req.IncludeMetadata {
		t.Error("Expected IncludeMetadata to be true")
	}

	if !req.IncludeContent {
		t.Error("Expected IncludeContent to be true")
	}
}

// TestSearchResult tests search result structure
func TestSearchResult(t *testing.T) {
	doc := &VectorDocument{
		ID:      "doc-1",
		Content: "Test content",
	}

	result := &SearchResult{
		Document: doc,
		Score:    0.95,
		Rank:     1,
	}

	if result.Document.ID != "doc-1" {
		t.Errorf("Expected document ID 'doc-1', got %s", result.Document.ID)
	}

	if result.Score != 0.95 {
		t.Errorf("Expected score 0.95, got %f", result.Score)
	}

	if result.Rank != 1 {
		t.Errorf("Expected rank 1, got %d", result.Rank)
	}
}

// TestVectorStoreError tests error handling
func TestVectorStoreError(t *testing.T) {
	err := NewVectorStoreError("Search", "test-collection",
		&HTTPError{StatusCode: 404, Message: "Collection not found"})

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	vsErr, ok := err.(*VectorStoreError)
	if !ok {
		t.Fatalf("Expected VectorStoreError, got %T", err)
	}

	if vsErr.Op != "Search" {
		t.Errorf("Expected operation 'Search', got %s", vsErr.Op)
	}

	if vsErr.Collection != "test-collection" {
		t.Errorf("Expected collection 'test-collection', got %s", vsErr.Collection)
	}

	expectedMsg := "vector store Search on collection 'test-collection': HTTP 404: Collection not found"
	if vsErr.Error() != expectedMsg {
		t.Errorf("Expected error message %q, got %q", expectedMsg, vsErr.Error())
	}
}

// TestVectorStoreErrorUnwrap tests error unwrapping
func TestVectorStoreErrorUnwrap(t *testing.T) {
	innerErr := &HTTPError{StatusCode: 500, Message: "Internal error"}
	err := NewVectorStoreError("Add", "test-collection", innerErr)

	vsErr, ok := err.(*VectorStoreError)
	if !ok {
		t.Fatalf("Expected VectorStoreError, got %T", err)
	}

	unwrapped := vsErr.Unwrap()
	if unwrapped != innerErr {
		t.Errorf("Expected unwrapped error to be innerErr")
	}
}

// MockVectorStore is a mock implementation for testing
type MockVectorStore struct {
	collections map[string][]*VectorDocument
}

func NewMockVectorStore() *MockVectorStore {
	return &MockVectorStore{
		collections: make(map[string][]*VectorDocument),
	}
}

func (m *MockVectorStore) CreateCollection(ctx context.Context, name string, config *CollectionConfig) error {
	if _, exists := m.collections[name]; exists {
		return NewVectorStoreError("CreateCollection", name,
			&HTTPError{StatusCode: 409, Message: "Collection already exists"})
	}
	m.collections[name] = make([]*VectorDocument, 0)
	return nil
}

func (m *MockVectorStore) DeleteCollection(ctx context.Context, name string) error {
	if _, exists := m.collections[name]; !exists {
		return NewVectorStoreError("DeleteCollection", name,
			&HTTPError{StatusCode: 404, Message: "Collection not found"})
	}
	delete(m.collections, name)
	return nil
}

func (m *MockVectorStore) ListCollections(ctx context.Context) ([]string, error) {
	names := make([]string, 0, len(m.collections))
	for name := range m.collections {
		names = append(names, name)
	}
	return names, nil
}

func (m *MockVectorStore) CollectionExists(ctx context.Context, name string) (bool, error) {
	_, exists := m.collections[name]
	return exists, nil
}

func (m *MockVectorStore) Add(ctx context.Context, collection string, docs []*VectorDocument) ([]string, error) {
	_, exists := m.collections[collection]
	if !exists {
		return nil, NewVectorStoreError("Add", collection,
			&HTTPError{StatusCode: 404, Message: "Collection not found"})
	}

	ids := make([]string, len(docs))
	for i, doc := range docs {
		if doc.ID == "" {
			doc.ID = "mock-" + time.Now().Format("20060102150405")
		}
		ids[i] = doc.ID
		m.collections[collection] = append(m.collections[collection], doc)
	}

	return ids, nil
}

func (m *MockVectorStore) Delete(ctx context.Context, collection string, ids []string) error {
	coll, exists := m.collections[collection]
	if !exists {
		return NewVectorStoreError("Delete", collection,
			&HTTPError{StatusCode: 404, Message: "Collection not found"})
	}

	// Remove documents with matching IDs
	filtered := make([]*VectorDocument, 0)
	for _, doc := range coll {
		found := false
		for _, id := range ids {
			if doc.ID == id {
				found = true
				break
			}
		}
		if !found {
			filtered = append(filtered, doc)
		}
	}

	m.collections[collection] = filtered
	return nil
}

func (m *MockVectorStore) Get(ctx context.Context, collection string, ids []string) ([]*VectorDocument, error) {
	coll, exists := m.collections[collection]
	if !exists {
		return nil, NewVectorStoreError("Get", collection,
			&HTTPError{StatusCode: 404, Message: "Collection not found"})
	}

	docs := make([]*VectorDocument, 0)
	for _, doc := range coll {
		for _, id := range ids {
			if doc.ID == id {
				docs = append(docs, doc)
				break
			}
		}
	}

	return docs, nil
}

func (m *MockVectorStore) Update(ctx context.Context, collection string, docs []*VectorDocument) error {
	_, err := m.Add(ctx, collection, docs)
	return err
}

func (m *MockVectorStore) Search(ctx context.Context, req *SearchRequest) ([]*SearchResult, error) {
	coll, exists := m.collections[req.Collection]
	if !exists {
		return nil, NewVectorStoreError("Search", req.Collection,
			&HTTPError{StatusCode: 404, Message: "Collection not found"})
	}

	// Simple mock: return first TopK documents with dummy scores
	results := make([]*SearchResult, 0)
	for i, doc := range coll {
		if i >= req.TopK {
			break
		}

		// Calculate dummy similarity based on position
		score := 1.0 - float32(i)*0.1

		if score < req.MinScore {
			continue
		}

		results = append(results, &SearchResult{
			Document: doc,
			Score:    score,
			Rank:     i + 1,
		})
	}

	return results, nil
}

func (m *MockVectorStore) SearchByText(ctx context.Context, req *TextSearchRequest) ([]*SearchResult, error) {
	// Mock: use dummy query vector
	searchReq := &SearchRequest{
		Collection:       req.Collection,
		QueryVector:      []float32{0.1, 0.2, 0.3},
		TopK:             req.TopK,
		Filter:           req.Filter,
		IncludeMetadata:  req.IncludeMetadata,
		IncludeContent:   req.IncludeContent,
		IncludeEmbedding: req.IncludeEmbedding,
		MinScore:         req.MinScore,
	}
	return m.Search(ctx, searchReq)
}

func (m *MockVectorStore) Count(ctx context.Context, collection string) (int64, error) {
	coll, exists := m.collections[collection]
	if !exists {
		return 0, NewVectorStoreError("Count", collection,
			&HTTPError{StatusCode: 404, Message: "Collection not found"})
	}
	return int64(len(coll)), nil
}

func (m *MockVectorStore) Clear(ctx context.Context, collection string) error {
	if _, exists := m.collections[collection]; !exists {
		return NewVectorStoreError("Clear", collection,
			&HTTPError{StatusCode: 404, Message: "Collection not found"})
	}
	m.collections[collection] = make([]*VectorDocument, 0)
	return nil
}

// TestMockVectorStore tests the mock implementation
func TestMockVectorStore(t *testing.T) {
	ctx := context.Background()
	store := NewMockVectorStore()

	// Test CreateCollection
	err := store.CreateCollection(ctx, "test", &CollectionConfig{
		Name:      "test",
		Dimension: 384,
	})
	if err != nil {
		t.Fatalf("CreateCollection failed: %v", err)
	}

	// Test duplicate collection creation
	err = store.CreateCollection(ctx, "test", nil)
	if err == nil {
		t.Error("Expected error for duplicate collection")
	}

	// Test CollectionExists
	exists, err := store.CollectionExists(ctx, "test")
	if err != nil {
		t.Fatalf("CollectionExists failed: %v", err)
	}
	if !exists {
		t.Error("Collection should exist")
	}

	// Test Add documents
	docs := []*VectorDocument{
		{
			ID:        "doc-1",
			Content:   "First document",
			Embedding: []float32{0.1, 0.2, 0.3},
		},
		{
			ID:        "doc-2",
			Content:   "Second document",
			Embedding: []float32{0.4, 0.5, 0.6},
		},
	}

	ids, err := store.Add(ctx, "test", docs)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	if len(ids) != 2 {
		t.Errorf("Expected 2 IDs, got %d", len(ids))
	}

	// Test Count
	count, err := store.Count(ctx, "test")
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}

	// Test Get
	retrieved, err := store.Get(ctx, "test", []string{"doc-1"})
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if len(retrieved) != 1 {
		t.Errorf("Expected 1 document, got %d", len(retrieved))
	}

	if retrieved[0].ID != "doc-1" {
		t.Errorf("Expected doc-1, got %s", retrieved[0].ID)
	}

	// Test Search
	searchReq := DefaultSearchRequest("test", []float32{0.1, 0.2, 0.3})
	searchReq.TopK = 1

	results, err := store.Search(ctx, searchReq)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	if results[0].Rank != 1 {
		t.Errorf("Expected rank 1, got %d", results[0].Rank)
	}

	// Test Delete
	err = store.Delete(ctx, "test", []string{"doc-1"})
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	count, _ = store.Count(ctx, "test")
	if count != 1 {
		t.Errorf("Expected count 1 after delete, got %d", count)
	}

	// Test Clear
	err = store.Clear(ctx, "test")
	if err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	count, _ = store.Count(ctx, "test")
	if count != 0 {
		t.Errorf("Expected count 0 after clear, got %d", count)
	}

	// Test ListCollections
	names, err := store.ListCollections(ctx)
	if err != nil {
		t.Fatalf("ListCollections failed: %v", err)
	}

	if len(names) != 1 {
		t.Errorf("Expected 1 collection, got %d", len(names))
	}

	// Test DeleteCollection
	err = store.DeleteCollection(ctx, "test")
	if err != nil {
		t.Fatalf("DeleteCollection failed: %v", err)
	}

	exists, _ = store.CollectionExists(ctx, "test")
	if exists {
		t.Error("Collection should not exist after deletion")
	}
}

// TestVectorStoreInterface verifies that implementations satisfy the interface
func TestVectorStoreInterface(t *testing.T) {
	var _ VectorStore = (*MockVectorStore)(nil)
	var _ VectorStore = (*ChromaStore)(nil)
}

// TestSearchRequestWithFilters tests search request with metadata filters
func TestSearchRequestWithFilters(t *testing.T) {
	req := DefaultSearchRequest("test", []float32{0.1, 0.2, 0.3})
	req.Filter = map[string]interface{}{
		"author": "John",
		"year":   2024,
	}

	if req.Filter["author"] != "John" {
		t.Errorf("Expected author 'John', got %v", req.Filter["author"])
	}

	if req.Filter["year"] != 2024 {
		t.Errorf("Expected year 2024, got %v", req.Filter["year"])
	}
}

// TestSearchResultRanking tests that results are properly ranked
func TestSearchResultRanking(t *testing.T) {
	results := []*SearchResult{
		{Score: 0.9, Rank: 1},
		{Score: 0.7, Rank: 2},
		{Score: 0.5, Rank: 3},
	}

	for i, result := range results {
		expectedRank := i + 1
		if result.Rank != expectedRank {
			t.Errorf("Expected rank %d, got %d", expectedRank, result.Rank)
		}

		// Verify scores are descending
		if i > 0 && result.Score > results[i-1].Score {
			t.Errorf("Results should be in descending order by score")
		}
	}
}
