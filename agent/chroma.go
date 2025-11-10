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
	// DefaultChromaURL is the default ChromaDB server URL
	DefaultChromaURL = "http://localhost:8000"

	// ChromaDB API version
	chromaAPIVersion = "v1"
)

// ChromaStore implements VectorStore for ChromaDB
type ChromaStore struct {
	baseURL   string
	client    *http.Client
	embedding EmbeddingProvider
}

// NewChromaStore creates a new ChromaDB vector store client
func NewChromaStore(baseURL string) (*ChromaStore, error) {
	if baseURL == "" {
		baseURL = DefaultChromaURL
	}

	return &ChromaStore{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// WithEmbedding sets the embedding provider for automatic embedding generation
func (c *ChromaStore) WithEmbedding(provider EmbeddingProvider) *ChromaStore {
	c.embedding = provider
	return c
}

// WithHTTPClient sets a custom HTTP client
func (c *ChromaStore) WithHTTPClient(client *http.Client) *ChromaStore {
	c.client = client
	return c
}

// chromaCollection represents a ChromaDB collection structure
type chromaCollection struct {
	Name     string                 `json:"name"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// chromaAddRequest represents the request structure for adding documents
type chromaAddRequest struct {
	IDs        []string                 `json:"ids"`
	Embeddings [][]float32              `json:"embeddings,omitempty"`
	Documents  []string                 `json:"documents,omitempty"`
	Metadatas  []map[string]interface{} `json:"metadatas,omitempty"`
}

// chromaQueryRequest represents the request structure for querying
type chromaQueryRequest struct {
	QueryEmbeddings [][]float32            `json:"query_embeddings"`
	NResults        int                    `json:"n_results"`
	Where           map[string]interface{} `json:"where,omitempty"`
	Include         []string               `json:"include,omitempty"`
}

// chromaQueryResponse represents the response from a query
type chromaQueryResponse struct {
	IDs        [][]string                 `json:"ids"`
	Embeddings [][][]float32              `json:"embeddings,omitempty"`
	Documents  [][]string                 `json:"documents,omitempty"`
	Metadatas  [][]map[string]interface{} `json:"metadatas,omitempty"`
	Distances  [][]float32                `json:"distances,omitempty"`
}

// CreateCollection creates a new ChromaDB collection
func (c *ChromaStore) CreateCollection(ctx context.Context, name string, config *CollectionConfig) error {
	metadata := make(map[string]interface{})

	if config != nil {
		if config.Description != "" {
			metadata["description"] = config.Description
		}
		if config.Dimension > 0 {
			metadata["dimension"] = config.Dimension
		}
		if config.DistanceMetric != "" {
			// ChromaDB uses "hnsw:space" for distance metric
			switch config.DistanceMetric {
			case DistanceMetricCosine:
				metadata["hnsw:space"] = "cosine"
			case DistanceMetricEuclidean, DistanceMetricL2:
				metadata["hnsw:space"] = "l2"
			case DistanceMetricDotProduct, DistanceMetricIP:
				metadata["hnsw:space"] = "ip"
			}
		}
	}

	reqBody := chromaCollection{
		Name:     name,
		Metadata: metadata,
	}

	_, err := c.doRequest(ctx, "POST", "/api/v1/collections", reqBody)
	if err != nil {
		return NewVectorStoreError("CreateCollection", name, err)
	}

	return nil
}

// DeleteCollection deletes a ChromaDB collection
func (c *ChromaStore) DeleteCollection(ctx context.Context, name string) error {
	_, err := c.doRequest(ctx, "DELETE", "/api/v1/collections/"+name, nil)
	if err != nil {
		return NewVectorStoreError("DeleteCollection", name, err)
	}
	return nil
}

// ListCollections returns all collection names
func (c *ChromaStore) ListCollections(ctx context.Context) ([]string, error) {
	resp, err := c.doRequest(ctx, "GET", "/api/v1/collections", nil)
	if err != nil {
		return nil, NewVectorStoreError("ListCollections", "", err)
	}

	var collections []chromaCollection
	if err := json.Unmarshal(resp, &collections); err != nil {
		return nil, NewVectorStoreError("ListCollections", "", err)
	}

	names := make([]string, len(collections))
	for i, coll := range collections {
		names[i] = coll.Name
	}

	return names, nil
}

// CollectionExists checks if a collection exists
func (c *ChromaStore) CollectionExists(ctx context.Context, name string) (bool, error) {
	_, err := c.doRequest(ctx, "GET", "/api/v1/collections/"+name, nil)
	if err != nil {
		// Check if it's a 404 error
		if httpErr, ok := err.(*HTTPError); ok && httpErr.StatusCode == http.StatusNotFound {
			return false, nil
		}
		return false, NewVectorStoreError("CollectionExists", name, err)
	}
	return true, nil
}

// Add inserts documents into a collection
func (c *ChromaStore) Add(ctx context.Context, collection string, docs []*VectorDocument) ([]string, error) {
	if len(docs) == 0 {
		return []string{}, nil
	}

	ids := make([]string, len(docs))
	embeddings := make([][]float32, len(docs))
	documents := make([]string, len(docs))
	metadatas := make([]map[string]interface{}, len(docs))

	for i, doc := range docs {
		// Use provided ID or generate one
		if doc.ID == "" {
			doc.ID = fmt.Sprintf("doc_%d_%d", time.Now().UnixNano(), i)
		}
		ids[i] = doc.ID
		documents[i] = doc.Content

		// Generate embedding if not provided
		if doc.Embedding == nil {
			if c.embedding == nil {
				return nil, NewVectorStoreError("Add", collection,
					fmt.Errorf("no embedding provided and no embedding provider configured\n\n"+
						"Fix:\n"+
						"  1. Set embedding provider: NewChromaStore(url).WithEmbedding(embedder)\n"+
						"  2. Or provide embeddings: doc.Embedding = []float32{...}\n"+
						"  3. Example: embedder, _ := agent.NewOpenAIEmbedding(\"text-embedding-3-small\", apiKey)\n"))
			}
			emb, err := c.embedding.Embed(ctx, doc.Content)
			if err != nil {
				return nil, NewVectorStoreError("Add", collection,
					fmt.Errorf("failed to generate embedding: %w\n\n"+
						"Possible causes:\n"+
						"  - Embedding API timeout (increase context deadline)\n"+
						"  - Document too long (max ~8,000 tokens)\n"+
						"  - Rate limit (add retry or use WithDefaults())\n"+
						"  - Invalid API key\n", err))
			}
			doc.Embedding = emb
		}
		embeddings[i] = doc.Embedding

		// Set metadata
		if doc.Metadata == nil {
			doc.Metadata = make(map[string]interface{})
		}
		doc.Metadata["created_at"] = doc.CreatedAt.Format(time.RFC3339)
		doc.Metadata["updated_at"] = doc.UpdatedAt.Format(time.RFC3339)
		metadatas[i] = doc.Metadata
	}

	reqBody := chromaAddRequest{
		IDs:        ids,
		Embeddings: embeddings,
		Documents:  documents,
		Metadatas:  metadatas,
	}

	_, err := c.doRequest(ctx, "POST", "/api/v1/collections/"+collection+"/add", reqBody)
	if err != nil {
		return nil, NewVectorStoreError("Add", collection, err)
	}

	return ids, nil
}

// Delete removes documents by IDs
func (c *ChromaStore) Delete(ctx context.Context, collection string, ids []string) error {
	reqBody := map[string]interface{}{
		"ids": ids,
	}

	_, err := c.doRequest(ctx, "POST", "/api/v1/collections/"+collection+"/delete", reqBody)
	if err != nil {
		return NewVectorStoreError("Delete", collection, err)
	}

	return nil
}

// Get retrieves documents by IDs
func (c *ChromaStore) Get(ctx context.Context, collection string, ids []string) ([]*VectorDocument, error) {
	reqBody := map[string]interface{}{
		"ids":     ids,
		"include": []string{"embeddings", "documents", "metadatas"},
	}

	resp, err := c.doRequest(ctx, "POST", "/api/v1/collections/"+collection+"/get", reqBody)
	if err != nil {
		return nil, NewVectorStoreError("Get", collection, err)
	}

	var result struct {
		IDs        []string                 `json:"ids"`
		Embeddings [][]float32              `json:"embeddings"`
		Documents  []string                 `json:"documents"`
		Metadatas  []map[string]interface{} `json:"metadatas"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, NewVectorStoreError("Get", collection, err)
	}

	docs := make([]*VectorDocument, len(result.IDs))
	for i := range result.IDs {
		doc := &VectorDocument{
			ID:       result.IDs[i],
			Metadata: make(map[string]interface{}),
		}

		if i < len(result.Documents) {
			doc.Content = result.Documents[i]
		}
		if i < len(result.Embeddings) {
			doc.Embedding = result.Embeddings[i]
		}
		if i < len(result.Metadatas) {
			doc.Metadata = result.Metadatas[i]

			// Parse timestamps
			if createdAt, ok := doc.Metadata["created_at"].(string); ok {
				if t, err := time.Parse(time.RFC3339, createdAt); err == nil {
					doc.CreatedAt = t
				}
			}
			if updatedAt, ok := doc.Metadata["updated_at"].(string); ok {
				if t, err := time.Parse(time.RFC3339, updatedAt); err == nil {
					doc.UpdatedAt = t
				}
			}
		}

		docs[i] = doc
	}

	return docs, nil
}

// Update updates documents by IDs (uses upsert)
func (c *ChromaStore) Update(ctx context.Context, collection string, docs []*VectorDocument) error {
	// ChromaDB uses upsert for updates
	_, err := c.Add(ctx, collection, docs)
	return err
}

// Search performs similarity search
func (c *ChromaStore) Search(ctx context.Context, req *SearchRequest) ([]*SearchResult, error) {
	include := []string{"distances"}
	if req.IncludeContent {
		include = append(include, "documents")
	}
	if req.IncludeMetadata {
		include = append(include, "metadatas")
	}
	if req.IncludeEmbedding {
		include = append(include, "embeddings")
	}

	queryReq := chromaQueryRequest{
		QueryEmbeddings: [][]float32{req.QueryVector},
		NResults:        req.TopK,
		Include:         include,
		Where:           req.Filter,
	}

	resp, err := c.doRequest(ctx, "POST", "/api/v1/collections/"+req.Collection+"/query", queryReq)
	if err != nil {
		return nil, NewVectorStoreError("Search", req.Collection, err)
	}

	var queryResp chromaQueryResponse
	if err := json.Unmarshal(resp, &queryResp); err != nil {
		return nil, NewVectorStoreError("Search", req.Collection, err)
	}

	// Parse results
	results := make([]*SearchResult, 0)
	if len(queryResp.IDs) > 0 && len(queryResp.IDs[0]) > 0 {
		for i, id := range queryResp.IDs[0] {
			doc := &VectorDocument{
				ID:       id,
				Metadata: make(map[string]interface{}),
			}

			// Get distance (lower is better in ChromaDB)
			var distance float32
			if len(queryResp.Distances) > 0 && i < len(queryResp.Distances[0]) {
				distance = queryResp.Distances[0][i]
			}

			// Convert distance to similarity score (higher is better)
			// For cosine: similarity = 1 - distance
			score := 1.0 - distance

			// Apply min score filter
			if score < req.MinScore {
				continue
			}

			// Get content
			if req.IncludeContent && len(queryResp.Documents) > 0 && i < len(queryResp.Documents[0]) {
				doc.Content = queryResp.Documents[0][i]
			}

			// Get metadata
			if req.IncludeMetadata && len(queryResp.Metadatas) > 0 && i < len(queryResp.Metadatas[0]) {
				doc.Metadata = queryResp.Metadatas[0][i]
			}

			// Get embedding
			if req.IncludeEmbedding && len(queryResp.Embeddings) > 0 && i < len(queryResp.Embeddings[0]) {
				doc.Embedding = queryResp.Embeddings[0][i]
			}

			results = append(results, &SearchResult{
				Document: doc,
				Score:    score,
				Rank:     i + 1,
			})
		}
	}

	return results, nil
}

// SearchByText generates embedding for text and performs search
func (c *ChromaStore) SearchByText(ctx context.Context, req *TextSearchRequest) ([]*SearchResult, error) {
	if c.embedding == nil {
		return nil, NewVectorStoreError("SearchByText", req.Collection,
			fmt.Errorf("no embedding provider configured"))
	}

	// Generate embedding for query text
	queryEmb, err := c.embedding.Embed(ctx, req.Query)
	if err != nil {
		return nil, NewVectorStoreError("SearchByText", req.Collection,
			fmt.Errorf("failed to generate query embedding: %w", err))
	}

	// Convert to SearchRequest and execute
	searchReq := &SearchRequest{
		Collection:       req.Collection,
		QueryVector:      queryEmb,
		TopK:             req.TopK,
		Filter:           req.Filter,
		IncludeMetadata:  req.IncludeMetadata,
		IncludeContent:   req.IncludeContent,
		IncludeEmbedding: req.IncludeEmbedding,
		MinScore:         req.MinScore,
	}

	return c.Search(ctx, searchReq)
}

// Count returns the number of documents in a collection
func (c *ChromaStore) Count(ctx context.Context, collection string) (int64, error) {
	resp, err := c.doRequest(ctx, "GET", "/api/v1/collections/"+collection+"/count", nil)
	if err != nil {
		return 0, NewVectorStoreError("Count", collection, err)
	}

	var count int64
	if err := json.Unmarshal(resp, &count); err != nil {
		return 0, NewVectorStoreError("Count", collection, err)
	}

	return count, nil
}

// Clear removes all documents from a collection
func (c *ChromaStore) Clear(ctx context.Context, collection string) error {
	// Get all document IDs first
	reqBody := map[string]interface{}{
		"include": []string{},
	}

	resp, err := c.doRequest(ctx, "POST", "/api/v1/collections/"+collection+"/get", reqBody)
	if err != nil {
		return NewVectorStoreError("Clear", collection, err)
	}

	var result struct {
		IDs []string `json:"ids"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return NewVectorStoreError("Clear", collection, err)
	}

	// Delete all documents
	if len(result.IDs) > 0 {
		return c.Delete(ctx, collection, result.IDs)
	}

	return nil
}

// HTTPError represents an HTTP error response
type HTTPError struct {
	StatusCode int
	Message    string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
}

// doRequest performs an HTTP request to ChromaDB
func (c *ChromaStore) doRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &HTTPError{
			StatusCode: resp.StatusCode,
			Message:    string(respBody),
		}
	}

	return respBody, nil
}
