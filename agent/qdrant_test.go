package agent

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewQdrantStore(t *testing.T) {
	store, err := NewQdrantStore("")
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	if store.baseURL != DefaultQdrantURL {
		t.Errorf("Expected baseURL %s, got %s", DefaultQdrantURL, store.baseURL)
	}

	customURL := "http://custom:6333"
	store, err = NewQdrantStore(customURL)
	if err != nil {
		t.Fatalf("Failed to create store with custom URL: %v", err)
	}

	if store.baseURL != customURL {
		t.Errorf("Expected baseURL %s, got %s", customURL, store.baseURL)
	}
}

func TestQdrantStoreWithMethods(t *testing.T) {
	store, _ := NewQdrantStore("")

	apiKey := "test-key"
	store.WithAPIKey(apiKey)
	if store.apiKey != apiKey {
		t.Errorf("Expected apiKey %s, got %s", apiKey, store.apiKey)
	}

	mockEmb := &MockEmbeddingProvider{
		embeddings: map[string][]float32{
			"test": {0.1, 0.2, 0.3},
		},
	}
	store.WithEmbedding(mockEmb)
	if store.embedding != mockEmb {
		t.Error("Expected embedding provider to be set")
	}

	customClient := &http.Client{Timeout: 10 * time.Second}
	store.WithHTTPClient(customClient)
	if store.client != customClient {
		t.Error("Expected HTTP client to be set")
	}
}

func TestQdrantCreateCollection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Expected PUT method, got %s", r.Method)
		}
		if r.URL.Path != "/collections/test-collection" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req qdrantCreateCollection
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
		}

		if req.Vectors.Size != 768 {
			t.Errorf("Expected dimension 768, got %d", req.Vectors.Size)
		}
		if req.Vectors.Distance != "Cosine" {
			t.Errorf("Expected distance Cosine, got %s", req.Vectors.Distance)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"result": true,
		})
	}))
	defer server.Close()

	store, _ := NewQdrantStore(server.URL)

	config := &CollectionConfig{
		Name:           "test-collection",
		Dimension:      768,
		DistanceMetric: DistanceMetricCosine,
	}

	err := store.CreateCollection(context.Background(), "test-collection", config)
	if err != nil {
		t.Errorf("Failed to create collection: %v", err)
	}
}

func TestQdrantCreateCollectionDistanceMetrics(t *testing.T) {
	tests := []struct {
		name           string
		metric         DistanceMetric
		expectedQdrant string
	}{
		{"Cosine", DistanceMetricCosine, "Cosine"},
		{"Euclidean", DistanceMetricEuclidean, "Euclid"},
		{"L2", DistanceMetricL2, "Euclid"},
		{"DotProduct", DistanceMetricDotProduct, "Dot"},
		{"IP", DistanceMetricIP, "Dot"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var req qdrantCreateCollection
				json.NewDecoder(r.Body).Decode(&req)

				if req.Vectors.Distance != tt.expectedQdrant {
					t.Errorf("Expected distance %s, got %s", tt.expectedQdrant, req.Vectors.Distance)
				}

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
			}))
			defer server.Close()

			store, _ := NewQdrantStore(server.URL)
			config := &CollectionConfig{
				Dimension:      128,
				DistanceMetric: tt.metric,
			}

			store.CreateCollection(context.Background(), "test", config)
		})
	}
}

func TestQdrantDeleteCollection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE method, got %s", r.Method)
		}
		if r.URL.Path != "/collections/test-collection" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"result": true,
		})
	}))
	defer server.Close()

	store, _ := NewQdrantStore(server.URL)
	err := store.DeleteCollection(context.Background(), "test-collection")
	if err != nil {
		t.Errorf("Failed to delete collection: %v", err)
	}
}

func TestQdrantListCollections(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET method, got %s", r.Method)
		}

		resp := map[string]interface{}{
			"status": "ok",
			"result": map[string]interface{}{
				"collections": []map[string]interface{}{
					{"name": "collection1"},
					{"name": "collection2"},
					{"name": "collection3"},
				},
			},
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	store, _ := NewQdrantStore(server.URL)
	collections, err := store.ListCollections(context.Background())
	if err != nil {
		t.Errorf("Failed to list collections: %v", err)
	}

	expected := []string{"collection1", "collection2", "collection3"}
	if len(collections) != len(expected) {
		t.Errorf("Expected %d collections, got %d", len(expected), len(collections))
	}

	for i, name := range expected {
		if collections[i] != name {
			t.Errorf("Expected collection %s, got %s", name, collections[i])
		}
	}
}

func TestQdrantCollectionExists(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/collections/exists" {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		} else {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"status": "error"})
		}
	}))
	defer server.Close()

	store, _ := NewQdrantStore(server.URL)

	exists, err := store.CollectionExists(context.Background(), "exists")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !exists {
		t.Error("Expected collection to exist")
	}

	exists, err = store.CollectionExists(context.Background(), "not-exists")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if exists {
		t.Error("Expected collection to not exist")
	}
}

func TestQdrantAdd(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Expected PUT method, got %s", r.Method)
		}

		var req qdrantUpsertRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
		}

		if len(req.Points) != 2 {
			t.Errorf("Expected 2 points, got %d", len(req.Points))
		}

		// Check first point
		if req.Points[0].ID != "doc1" {
			t.Errorf("Expected ID doc1, got %v", req.Points[0].ID)
		}
		if content, ok := req.Points[0].Payload["content"].(string); !ok || content != "Test content 1" {
			t.Errorf("Unexpected content in payload")
		}
		if len(req.Points[0].Vector) != 3 {
			t.Errorf("Expected 3-dim vector, got %d", len(req.Points[0].Vector))
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"result": map[string]interface{}{
				"operation_id": 1,
				"status":       "completed",
			},
		})
	}))
	defer server.Close()

	store, _ := NewQdrantStore(server.URL)

	docs := []*VectorDocument{
		{
			ID:        "doc1",
			Content:   "Test content 1",
			Embedding: []float32{0.1, 0.2, 0.3},
			Metadata: map[string]interface{}{
				"category": "test",
			},
		},
		{
			ID:        "doc2",
			Content:   "Test content 2",
			Embedding: []float32{0.4, 0.5, 0.6},
		},
	}

	ids, err := store.Add(context.Background(), "test-collection", docs)
	if err != nil {
		t.Errorf("Failed to add documents: %v", err)
	}

	if len(ids) != 2 {
		t.Errorf("Expected 2 IDs, got %d", len(ids))
	}
	if ids[0] != "doc1" || ids[1] != "doc2" {
		t.Errorf("Unexpected IDs returned: %v", ids)
	}
}

func TestQdrantAddWithAutoEmbedding(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	store, _ := NewQdrantStore(server.URL)
	mockEmb := &MockEmbeddingProvider{
		embeddings: map[string][]float32{
			"Auto embed me": {0.1, 0.2, 0.3},
		},
	}
	store.WithEmbedding(mockEmb)

	docs := []*VectorDocument{
		{
			ID:      "doc1",
			Content: "Auto embed me",
		},
	}

	_, err := store.Add(context.Background(), "test", docs)
	if err != nil {
		t.Errorf("Failed to add with auto-embedding: %v", err)
	}

	if docs[0].Embedding == nil {
		t.Error("Expected embedding to be generated")
	}
}

func TestQdrantAddNoEmbedding(t *testing.T) {
	store, _ := NewQdrantStore("")

	docs := []*VectorDocument{
		{
			ID:      "doc1",
			Content: "No embedding",
		},
	}

	_, err := store.Add(context.Background(), "test", docs)
	if err == nil {
		t.Error("Expected error when no embedding provided")
	}
}

func TestQdrantDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		var req map[string]interface{}
		json.NewDecoder(r.Body).Decode(&req)

		points, ok := req["points"].([]interface{})
		if !ok || len(points) != 2 {
			t.Errorf("Expected 2 point IDs to delete")
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	store, _ := NewQdrantStore(server.URL)
	err := store.Delete(context.Background(), "test", []string{"doc1", "doc2"})
	if err != nil {
		t.Errorf("Failed to delete: %v", err)
	}
}

func TestQdrantGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"status": "ok",
			"result": []map[string]interface{}{
				{
					"id":     "doc1",
					"vector": []float32{0.1, 0.2, 0.3},
					"payload": map[string]interface{}{
						"content":    "Test content",
						"category":   "test",
						"created_at": "2024-01-01T00:00:00Z",
					},
				},
				{
					"id":     float64(123), // Test numeric ID
					"vector": []float32{0.4, 0.5, 0.6},
					"payload": map[string]interface{}{
						"content": "Another test",
					},
				},
			},
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	store, _ := NewQdrantStore(server.URL)
	docs, err := store.Get(context.Background(), "test", []string{"doc1", "123"})
	if err != nil {
		t.Errorf("Failed to get documents: %v", err)
	}

	if len(docs) != 2 {
		t.Fatalf("Expected 2 documents, got %d", len(docs))
	}

	// Check first doc
	if docs[0].ID != "doc1" {
		t.Errorf("Expected ID doc1, got %s", docs[0].ID)
	}
	if docs[0].Content != "Test content" {
		t.Errorf("Expected content 'Test content', got %s", docs[0].Content)
	}
	if docs[0].Metadata["category"] != "test" {
		t.Errorf("Expected category metadata")
	}
	if docs[0].CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be parsed")
	}

	// Check second doc with numeric ID
	if docs[1].ID != "123" {
		t.Errorf("Expected ID 123, got %s", docs[1].ID)
	}
}

func TestQdrantSearch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		var req qdrantSearchRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.Limit != 5 {
			t.Errorf("Expected limit 5, got %d", req.Limit)
		}
		if len(req.Vector) != 3 {
			t.Errorf("Expected 3-dim query vector")
		}

		resp := map[string]interface{}{
			"status": "ok",
			"result": []map[string]interface{}{
				{
					"id":    "doc1",
					"score": 0.95,
					"payload": map[string]interface{}{
						"content": "Most relevant",
					},
					"vector": []float32{0.1, 0.2, 0.3},
				},
				{
					"id":    "doc2",
					"score": 0.85,
					"payload": map[string]interface{}{
						"content": "Less relevant",
					},
				},
			},
			"time": 0.005,
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	store, _ := NewQdrantStore(server.URL)

	searchReq := &SearchRequest{
		Collection:       "test",
		QueryVector:      []float32{0.1, 0.2, 0.3},
		TopK:             5,
		IncludeContent:   true,
		IncludeMetadata:  true,
		IncludeEmbedding: true,
	}

	results, err := store.Search(context.Background(), searchReq)
	if err != nil {
		t.Errorf("Failed to search: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}

	// Check first result
	if results[0].Document.ID != "doc1" {
		t.Errorf("Expected ID doc1, got %s", results[0].Document.ID)
	}
	if results[0].Score != 0.95 {
		t.Errorf("Expected score 0.95, got %f", results[0].Score)
	}
	if results[0].Rank != 1 {
		t.Errorf("Expected rank 1, got %d", results[0].Rank)
	}
	if results[0].Document.Content != "Most relevant" {
		t.Errorf("Unexpected content")
	}
	if results[0].Document.Embedding == nil {
		t.Error("Expected embedding to be included")
	}

	// Check second result
	if results[1].Rank != 2 {
		t.Errorf("Expected rank 2, got %d", results[1].Rank)
	}
}

func TestQdrantSearchWithFilter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req qdrantSearchRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.Filter == nil {
			t.Error("Expected filter to be set")
		}

		must, ok := req.Filter["must"]
		if !ok {
			t.Error("Expected must conditions in filter")
		}

		// Check must is a slice (can be []interface{} after JSON marshaling)
		mustSlice, ok := must.([]interface{})
		if !ok {
			t.Errorf("Expected must to be a slice, got %T", must)
		} else if len(mustSlice) == 0 {
			t.Error("Expected non-empty must conditions")
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"result": []interface{}{},
			"time":   0.001,
		})
	}))
	defer server.Close()

	store, _ := NewQdrantStore(server.URL)

	searchReq := &SearchRequest{
		Collection:  "test",
		QueryVector: []float32{0.1, 0.2, 0.3},
		TopK:        10,
		Filter: map[string]interface{}{
			"category": "science",
			"year":     2024,
		},
	}

	_, err := store.Search(context.Background(), searchReq)
	if err != nil {
		t.Errorf("Failed to search with filter: %v", err)
	}
}

func TestQdrantSearchByText(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"result": []interface{}{},
			"time":   0.001,
		})
	}))
	defer server.Close()

	store, _ := NewQdrantStore(server.URL)
	mockEmb := &MockEmbeddingProvider{
		embeddings: map[string][]float32{
			"search query": {0.1, 0.2, 0.3},
		},
	}
	store.WithEmbedding(mockEmb)

	textReq := &TextSearchRequest{
		Collection: "test",
		Query:      "search query",
		TopK:       10,
	}

	_, err := store.SearchByText(context.Background(), textReq)
	if err != nil {
		t.Errorf("Failed to search by text: %v", err)
	}
}

func TestQdrantSearchByTextNoProvider(t *testing.T) {
	store, _ := NewQdrantStore("")

	textReq := &TextSearchRequest{
		Collection: "test",
		Query:      "search query",
	}

	_, err := store.SearchByText(context.Background(), textReq)
	if err == nil {
		t.Error("Expected error when no embedding provider")
	}
}

func TestQdrantCount(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"status": "ok",
			"result": map[string]interface{}{
				"points_count": 42,
			},
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	store, _ := NewQdrantStore(server.URL)
	count, err := store.Count(context.Background(), "test")
	if err != nil {
		t.Errorf("Failed to get count: %v", err)
	}

	if count != 42 {
		t.Errorf("Expected count 42, got %d", count)
	}
}

func TestQdrantClear(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		// First call: GET collection info
		if r.Method == "GET" && callCount == 1 {
			resp := map[string]interface{}{
				"status": "ok",
				"result": map[string]interface{}{
					"status": "green",
					"config": map[string]interface{}{
						"params": map[string]interface{}{
							"vectors": map[string]interface{}{
								"size":     768,
								"distance": "Cosine",
							},
						},
					},
				},
			}
			json.NewEncoder(w).Encode(resp)
			return
		}

		// Second call: DELETE collection
		if r.Method == "DELETE" && callCount == 2 {
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
			return
		}

		// Third call: PUT (recreate) collection
		if r.Method == "PUT" && callCount == 3 {
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
			return
		}

		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	store, _ := NewQdrantStore(server.URL)
	err := store.Clear(context.Background(), "test")
	if err != nil {
		t.Errorf("Failed to clear collection: %v", err)
	}

	if callCount != 3 {
		t.Errorf("Expected 3 API calls (GET, DELETE, PUT), got %d", callCount)
	}
}

func TestQdrantUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Expected PUT method for update, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	store, _ := NewQdrantStore(server.URL)

	docs := []*VectorDocument{
		{
			ID:        "doc1",
			Content:   "Updated content",
			Embedding: []float32{0.7, 0.8, 0.9},
		},
	}

	err := store.Update(context.Background(), "test", docs)
	if err != nil {
		t.Errorf("Failed to update: %v", err)
	}
}

func TestQdrantAPIKey(t *testing.T) {
	apiKeySent := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKeySent = r.Header.Get("api-key")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	store, _ := NewQdrantStore(server.URL)
	store.WithAPIKey("secret-key")

	store.DeleteCollection(context.Background(), "test")

	if apiKeySent != "secret-key" {
		t.Errorf("Expected API key 'secret-key', got '%s'", apiKeySent)
	}
}

func TestConvertFilterToQdrant(t *testing.T) {
	filter := map[string]interface{}{
		"category": "science",
		"year":     2024,
		"active":   true,
	}

	qdrantFilter := convertFilterToQdrant(filter)

	must, ok := qdrantFilter["must"].([]map[string]interface{})
	if !ok {
		t.Fatal("Expected must field in filter")
	}

	if len(must) != 3 {
		t.Errorf("Expected 3 conditions, got %d", len(must))
	}

	// Verify structure
	foundCategory := false
	for _, condition := range must {
		if condition["key"] == "category" {
			foundCategory = true
			match, ok := condition["match"].(map[string]interface{})
			if !ok {
				t.Error("Expected match field")
			}
			if match["value"] != "science" {
				t.Error("Expected value 'science'")
			}
		}
	}

	if !foundCategory {
		t.Error("Expected category condition in filter")
	}
}

func TestQdrantStoreImplementsVectorStore(t *testing.T) {
	var _ VectorStore = (*QdrantStore)(nil)
}
