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
	// DefaultQdrantURL is the default Qdrant server URL
	DefaultQdrantURL = "http://localhost:6333"
)

// QdrantStore implements VectorStore for Qdrant
type QdrantStore struct {
	baseURL   string
	apiKey    string
	client    *http.Client
	embedding EmbeddingProvider
}

// NewQdrantStore creates a new Qdrant vector store client
func NewQdrantStore(baseURL string) (*QdrantStore, error) {
	if baseURL == "" {
		baseURL = DefaultQdrantURL
	}

	return &QdrantStore{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// WithAPIKey sets the API key for authentication
func (q *QdrantStore) WithAPIKey(apiKey string) *QdrantStore {
	q.apiKey = apiKey
	return q
}

// WithEmbedding sets the embedding provider for automatic embedding generation
func (q *QdrantStore) WithEmbedding(provider EmbeddingProvider) *QdrantStore {
	q.embedding = provider
	return q
}

// WithHTTPClient sets a custom HTTP client
func (q *QdrantStore) WithHTTPClient(client *http.Client) *QdrantStore {
	q.client = client
	return q
}

// Qdrant API structures

// qdrantCollectionInfo represents collection information
type qdrantCollectionInfo struct {
	Status string `json:"status"`
	Result struct {
		Status string `json:"status"`
		Config struct {
			Params struct {
				Vectors struct {
					Size     int    `json:"size"`
					Distance string `json:"distance"`
				} `json:"vectors"`
			} `json:"params"`
		} `json:"config"`
	} `json:"result,omitempty"`
}

// qdrantCreateCollection represents collection creation request
type qdrantCreateCollection struct {
	Vectors struct {
		Size     int    `json:"size"`
		Distance string `json:"distance"`
	} `json:"vectors"`
}

// qdrantPoint represents a point in Qdrant
type qdrantPoint struct {
	ID      interface{}            `json:"id"`
	Vector  []float32              `json:"vector"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}

// qdrantUpsertRequest represents the upsert request
type qdrantUpsertRequest struct {
	Points []qdrantPoint `json:"points"`
}

// qdrantSearchRequest represents the search request
type qdrantSearchRequest struct {
	Vector         []float32              `json:"vector"`
	Limit          int                    `json:"limit"`
	WithPayload    bool                   `json:"with_payload"`
	WithVector     bool                   `json:"with_vector"`
	Filter         map[string]interface{} `json:"filter,omitempty"`
	ScoreThreshold *float32               `json:"score_threshold,omitempty"`
}

// qdrantSearchResponse represents the search response
type qdrantSearchResponse struct {
	Result []struct {
		ID      interface{}            `json:"id"`
		Version int                    `json:"version"`
		Score   float32                `json:"score"`
		Payload map[string]interface{} `json:"payload,omitempty"`
		Vector  []float32              `json:"vector,omitempty"`
	} `json:"result"`
	Status string  `json:"status"`
	Time   float64 `json:"time"`
}

// qdrantScrollRequest represents scroll pagination request
type qdrantScrollRequest struct {
	Limit       int                    `json:"limit"`
	WithPayload bool                   `json:"with_payload"`
	WithVector  bool                   `json:"with_vector"`
	Filter      map[string]interface{} `json:"filter,omitempty"`
	Offset      interface{}            `json:"offset,omitempty"`
}

// qdrantScrollResponse represents scroll pagination response
type qdrantScrollResponse struct {
	Result struct {
		Points []struct {
			ID      interface{}            `json:"id"`
			Version int                    `json:"version"`
			Payload map[string]interface{} `json:"payload,omitempty"`
			Vector  []float32              `json:"vector,omitempty"`
		} `json:"points"`
		NextPageOffset interface{} `json:"next_page_offset"`
	} `json:"result"`
	Status string  `json:"status"`
	Time   float64 `json:"time"`
}

// CreateCollection creates a new Qdrant collection
func (q *QdrantStore) CreateCollection(ctx context.Context, name string, config *CollectionConfig) error {
	if config == nil || config.Dimension == 0 {
		return NewVectorStoreError("CreateCollection", name,
			fmt.Errorf("dimension is required"))
	}

	// Map distance metric
	distance := "Cosine"
	if config.DistanceMetric != "" {
		switch config.DistanceMetric {
		case DistanceMetricCosine:
			distance = "Cosine"
		case DistanceMetricEuclidean, DistanceMetricL2:
			distance = "Euclid"
		case DistanceMetricDotProduct, DistanceMetricIP:
			distance = "Dot"
		}
	}

	reqBody := qdrantCreateCollection{
		Vectors: struct {
			Size     int    `json:"size"`
			Distance string `json:"distance"`
		}{
			Size:     config.Dimension,
			Distance: distance,
		},
	}

	_, err := q.doRequest(ctx, "PUT", "/collections/"+name, reqBody)
	if err != nil {
		return NewVectorStoreError("CreateCollection", name, err)
	}

	return nil
}

// DeleteCollection deletes a Qdrant collection
func (q *QdrantStore) DeleteCollection(ctx context.Context, name string) error {
	_, err := q.doRequest(ctx, "DELETE", "/collections/"+name, nil)
	if err != nil {
		return NewVectorStoreError("DeleteCollection", name, err)
	}
	return nil
}

// ListCollections returns all collection names
func (q *QdrantStore) ListCollections(ctx context.Context) ([]string, error) {
	resp, err := q.doRequest(ctx, "GET", "/collections", nil)
	if err != nil {
		return nil, NewVectorStoreError("ListCollections", "", err)
	}

	var result struct {
		Result struct {
			Collections []struct {
				Name string `json:"name"`
			} `json:"collections"`
		} `json:"result"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, NewVectorStoreError("ListCollections", "", err)
	}

	names := make([]string, len(result.Result.Collections))
	for i, coll := range result.Result.Collections {
		names[i] = coll.Name
	}

	return names, nil
}

// CollectionExists checks if a collection exists
func (q *QdrantStore) CollectionExists(ctx context.Context, name string) (bool, error) {
	_, err := q.doRequest(ctx, "GET", "/collections/"+name, nil)
	if err != nil {
		if httpErr, ok := err.(*HTTPError); ok && httpErr.StatusCode == http.StatusNotFound {
			return false, nil
		}
		return false, NewVectorStoreError("CollectionExists", name, err)
	}
	return true, nil
}

// Add inserts documents into a collection
func (q *QdrantStore) Add(ctx context.Context, collection string, docs []*VectorDocument) ([]string, error) {
	if len(docs) == 0 {
		return []string{}, nil
	}

	points := make([]qdrantPoint, len(docs))
	ids := make([]string, len(docs))

	for i, doc := range docs {
		// Use provided ID or generate one
		if doc.ID == "" {
			doc.ID = fmt.Sprintf("point_%d_%d", time.Now().UnixNano(), i)
		}
		ids[i] = doc.ID

		// Generate embedding if not provided
		if doc.Embedding == nil {
			if q.embedding == nil {
				return nil, NewVectorStoreError("Add", collection,
					fmt.Errorf("no embedding provided and no embedding provider configured"))
			}
			emb, err := q.embedding.Embed(ctx, doc.Content)
			if err != nil {
				return nil, NewVectorStoreError("Add", collection,
					fmt.Errorf("failed to generate embedding: %w", err))
			}
			doc.Embedding = emb
		}

		// Prepare payload
		payload := make(map[string]interface{})
		if doc.Content != "" {
			payload["content"] = doc.Content
		}
		if doc.Metadata != nil {
			for k, v := range doc.Metadata {
				payload[k] = v
			}
		}
		if !doc.CreatedAt.IsZero() {
			payload["created_at"] = doc.CreatedAt.Format(time.RFC3339)
		}
		if !doc.UpdatedAt.IsZero() {
			payload["updated_at"] = doc.UpdatedAt.Format(time.RFC3339)
		}

		points[i] = qdrantPoint{
			ID:      doc.ID,
			Vector:  doc.Embedding,
			Payload: payload,
		}
	}

	reqBody := qdrantUpsertRequest{
		Points: points,
	}

	_, err := q.doRequest(ctx, "PUT", "/collections/"+collection+"/points", reqBody)
	if err != nil {
		return nil, NewVectorStoreError("Add", collection, err)
	}

	return ids, nil
}

// Delete removes documents by IDs
func (q *QdrantStore) Delete(ctx context.Context, collection string, ids []string) error {
	reqBody := map[string]interface{}{
		"points": ids,
	}

	_, err := q.doRequest(ctx, "POST", "/collections/"+collection+"/points/delete", reqBody)
	if err != nil {
		return NewVectorStoreError("Delete", collection, err)
	}

	return nil
}

// Get retrieves documents by IDs
func (q *QdrantStore) Get(ctx context.Context, collection string, ids []string) ([]*VectorDocument, error) {
	reqBody := map[string]interface{}{
		"ids":          ids,
		"with_payload": true,
		"with_vector":  true,
	}

	resp, err := q.doRequest(ctx, "POST", "/collections/"+collection+"/points", reqBody)
	if err != nil {
		return nil, NewVectorStoreError("Get", collection, err)
	}

	var result struct {
		Result []struct {
			ID      interface{}            `json:"id"`
			Payload map[string]interface{} `json:"payload"`
			Vector  []float32              `json:"vector"`
		} `json:"result"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, NewVectorStoreError("Get", collection, err)
	}

	docs := make([]*VectorDocument, len(result.Result))
	for i, point := range result.Result {
		doc := &VectorDocument{
			Embedding: point.Vector,
			Metadata:  make(map[string]interface{}),
		}

		// Extract ID as string
		switch v := point.ID.(type) {
		case string:
			doc.ID = v
		case float64:
			doc.ID = fmt.Sprintf("%.0f", v)
		default:
			doc.ID = fmt.Sprintf("%v", v)
		}

		// Extract content and metadata from payload
		if point.Payload != nil {
			if content, ok := point.Payload["content"].(string); ok {
				doc.Content = content
			}

			for k, v := range point.Payload {
				if k == "content" {
					continue
				}
				if k == "created_at" {
					if ts, ok := v.(string); ok {
						if t, err := time.Parse(time.RFC3339, ts); err == nil {
							doc.CreatedAt = t
						}
					}
					continue
				}
				if k == "updated_at" {
					if ts, ok := v.(string); ok {
						if t, err := time.Parse(time.RFC3339, ts); err == nil {
							doc.UpdatedAt = t
						}
					}
					continue
				}
				doc.Metadata[k] = v
			}
		}

		docs[i] = doc
	}

	return docs, nil
}

// Update updates documents by IDs (uses upsert)
func (q *QdrantStore) Update(ctx context.Context, collection string, docs []*VectorDocument) error {
	_, err := q.Add(ctx, collection, docs)
	return err
}

// Search performs similarity search
func (q *QdrantStore) Search(ctx context.Context, req *SearchRequest) ([]*SearchResult, error) {
	searchReq := qdrantSearchRequest{
		Vector:      req.QueryVector,
		Limit:       req.TopK,
		WithPayload: req.IncludeContent || req.IncludeMetadata,
		WithVector:  req.IncludeEmbedding,
	}

	// Add score threshold
	if req.MinScore > 0 {
		searchReq.ScoreThreshold = &req.MinScore
	}

	// Add filter if provided
	if req.Filter != nil && len(req.Filter) > 0 {
		searchReq.Filter = convertFilterToQdrant(req.Filter)
	}

	resp, err := q.doRequest(ctx, "POST", "/collections/"+req.Collection+"/points/search", searchReq)
	if err != nil {
		return nil, NewVectorStoreError("Search", req.Collection, err)
	}

	var searchResp qdrantSearchResponse
	if err := json.Unmarshal(resp, &searchResp); err != nil {
		return nil, NewVectorStoreError("Search", req.Collection, err)
	}

	results := make([]*SearchResult, 0, len(searchResp.Result))
	for i, point := range searchResp.Result {
		doc := &VectorDocument{
			Metadata: make(map[string]interface{}),
		}

		// Extract ID
		switch v := point.ID.(type) {
		case string:
			doc.ID = v
		case float64:
			doc.ID = fmt.Sprintf("%.0f", v)
		default:
			doc.ID = fmt.Sprintf("%v", v)
		}

		// Extract content and metadata from payload
		if point.Payload != nil {
			if content, ok := point.Payload["content"].(string); ok {
				doc.Content = content
			}

			for k, v := range point.Payload {
				if k != "content" && k != "created_at" && k != "updated_at" {
					doc.Metadata[k] = v
				}
			}
		}

		// Include vector if requested
		if req.IncludeEmbedding {
			doc.Embedding = point.Vector
		}

		results = append(results, &SearchResult{
			Document: doc,
			Score:    point.Score,
			Rank:     i + 1,
		})
	}

	return results, nil
}

// SearchByText generates embedding for text and performs search
func (q *QdrantStore) SearchByText(ctx context.Context, req *TextSearchRequest) ([]*SearchResult, error) {
	if q.embedding == nil {
		return nil, NewVectorStoreError("SearchByText", req.Collection,
			fmt.Errorf("no embedding provider configured"))
	}

	// Generate embedding for query text
	queryEmb, err := q.embedding.Embed(ctx, req.Query)
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

	return q.Search(ctx, searchReq)
}

// Count returns the number of documents in a collection
func (q *QdrantStore) Count(ctx context.Context, collection string) (int64, error) {
	resp, err := q.doRequest(ctx, "GET", "/collections/"+collection, nil)
	if err != nil {
		return 0, NewVectorStoreError("Count", collection, err)
	}

	var result struct {
		Result struct {
			PointsCount int64 `json:"points_count"`
		} `json:"result"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return 0, NewVectorStoreError("Count", collection, err)
	}

	return result.Result.PointsCount, nil
}

// Clear removes all documents from a collection
func (q *QdrantStore) Clear(ctx context.Context, collection string) error {
	// In Qdrant, we need to scroll through and delete all points
	// For simplicity, we'll recreate the collection
	// First get collection info
	resp, err := q.doRequest(ctx, "GET", "/collections/"+collection, nil)
	if err != nil {
		return NewVectorStoreError("Clear", collection, err)
	}

	var collInfo qdrantCollectionInfo
	if err := json.Unmarshal(resp, &collInfo); err != nil {
		return NewVectorStoreError("Clear", collection, err)
	}

	// Delete collection
	if err := q.DeleteCollection(ctx, collection); err != nil {
		return err
	}

	// Recreate with same config
	config := &CollectionConfig{
		Dimension: collInfo.Result.Config.Params.Vectors.Size,
	}

	// Map distance back
	switch collInfo.Result.Config.Params.Vectors.Distance {
	case "Cosine":
		config.DistanceMetric = DistanceMetricCosine
	case "Euclid":
		config.DistanceMetric = DistanceMetricEuclidean
	case "Dot":
		config.DistanceMetric = DistanceMetricDotProduct
	}

	return q.CreateCollection(ctx, collection, config)
}

// convertFilterToQdrant converts simple filter to Qdrant filter format
func convertFilterToQdrant(filter map[string]interface{}) map[string]interface{} {
	must := make([]map[string]interface{}, 0)

	for key, value := range filter {
		must = append(must, map[string]interface{}{
			"key": key,
			"match": map[string]interface{}{
				"value": value,
			},
		})
	}

	return map[string]interface{}{
		"must": must,
	}
}

// doRequest performs an HTTP request to Qdrant
func (q *QdrantStore) doRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	url := q.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Add API key if configured
	if q.apiKey != "" {
		req.Header.Set("api-key", q.apiKey)
	}

	resp, err := q.client.Do(req)
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
