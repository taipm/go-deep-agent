package agent

import (
	"context"
	"time"
)

// VectorStore is the interface for vector database operations
// Implementations include ChromaDB, Qdrant, Weaviate, Pinecone, etc.
type VectorStore interface {
	// Collection Operations

	// CreateCollection creates a new collection with the specified name and configuration
	CreateCollection(ctx context.Context, name string, config *CollectionConfig) error

	// DeleteCollection deletes a collection by name
	DeleteCollection(ctx context.Context, name string) error

	// ListCollections returns all collection names
	ListCollections(ctx context.Context) ([]string, error)

	// CollectionExists checks if a collection exists
	CollectionExists(ctx context.Context, name string) (bool, error)

	// Document Operations

	// Add inserts or updates documents with their embeddings and metadata
	// Returns the IDs of the added documents
	Add(ctx context.Context, collection string, docs []*VectorDocument) ([]string, error)

	// Delete removes documents by their IDs
	Delete(ctx context.Context, collection string, ids []string) error

	// Get retrieves documents by their IDs
	Get(ctx context.Context, collection string, ids []string) ([]*VectorDocument, error)

	// Update updates documents by their IDs
	Update(ctx context.Context, collection string, docs []*VectorDocument) error

	// Search Operations

	// Search performs similarity search with optional metadata filtering
	// Returns the top-k most similar documents
	Search(ctx context.Context, req *SearchRequest) ([]*SearchResult, error)

	// SearchByText generates embedding for text and performs similarity search
	// Requires an embedding provider to be configured
	SearchByText(ctx context.Context, req *TextSearchRequest) ([]*SearchResult, error)

	// Utility Operations

	// Count returns the number of documents in a collection
	Count(ctx context.Context, collection string) (int64, error)

	// Clear removes all documents from a collection (keeps collection)
	Clear(ctx context.Context, collection string) error
}

// VectorDocument represents a document with embeddings and metadata in vector store
type VectorDocument struct {
	// ID is the unique identifier for the document
	ID string

	// Content is the text content of the document
	Content string

	// Embedding is the vector representation of the content
	// If nil, the vector store should generate it using a configured embedding provider
	Embedding []float32

	// Metadata contains additional information about the document
	// Examples: {"source": "web", "author": "John", "timestamp": "2024-01-01"}
	Metadata map[string]interface{}

	// CreatedAt is the document creation timestamp
	CreatedAt time.Time

	// UpdatedAt is the last update timestamp
	UpdatedAt time.Time
}

// CollectionConfig contains configuration for creating a collection
type CollectionConfig struct {
	// Name of the collection
	Name string

	// Description of the collection's purpose
	Description string

	// Dimension of the embedding vectors
	// Must match the embedding provider's output dimension
	Dimension int

	// DistanceMetric specifies how to calculate similarity
	// Options: "cosine", "euclidean", "dot_product"
	DistanceMetric DistanceMetric

	// EmbeddingProvider is used to generate embeddings for documents
	// If nil, embeddings must be provided explicitly in Document.Embedding
	EmbeddingProvider EmbeddingProvider

	// Metadata schema for validation (optional)
	// Map of field name to field type: {"author": "string", "year": "int"}
	MetadataSchema map[string]string
}

// DistanceMetric defines how similarity is calculated
type DistanceMetric string

const (
	// DistanceMetricCosine measures cosine similarity (angular distance)
	// Range: -1 to 1, where 1 is most similar
	// Best for: normalized vectors, semantic similarity
	DistanceMetricCosine DistanceMetric = "cosine"

	// DistanceMetricEuclidean measures L2 distance
	// Range: 0 to infinity, where 0 is most similar
	// Best for: spatial data, absolute distances
	DistanceMetricEuclidean DistanceMetric = "euclidean"

	// DistanceMetricDotProduct measures inner product
	// Range: -infinity to infinity, where higher is more similar
	// Best for: when magnitude matters
	DistanceMetricDotProduct DistanceMetric = "dot_product"

	// DistanceMetricL2 is an alias for Euclidean
	DistanceMetricL2 DistanceMetric = "l2"

	// DistanceMetricIP is an alias for inner product (dot product)
	DistanceMetricIP DistanceMetric = "ip"
)

// SearchRequest contains parameters for similarity search
type SearchRequest struct {
	// Collection to search in
	Collection string

	// QueryVector is the embedding to search for
	QueryVector []float32

	// TopK is the number of results to return
	TopK int

	// Filter is metadata filtering criteria
	// Examples: {"author": "John"}, {"year": {"$gte": 2020}}
	Filter map[string]interface{}

	// IncludeMetadata specifies whether to include metadata in results
	IncludeMetadata bool

	// IncludeContent specifies whether to include document content in results
	IncludeContent bool

	// IncludeEmbedding specifies whether to include embeddings in results
	IncludeEmbedding bool

	// MinScore is the minimum similarity score to return
	// Documents with score < MinScore are excluded
	MinScore float32
}

// TextSearchRequest contains parameters for text-based search
// The vector store will use its embedding provider to generate the query vector
type TextSearchRequest struct {
	// Collection to search in
	Collection string

	// Query is the text to search for
	Query string

	// TopK is the number of results to return
	TopK int

	// Filter is metadata filtering criteria
	Filter map[string]interface{}

	// IncludeMetadata specifies whether to include metadata in results
	IncludeMetadata bool

	// IncludeContent specifies whether to include document content in results
	IncludeContent bool

	// IncludeEmbedding specifies whether to include embeddings in results
	IncludeEmbedding bool

	// MinScore is the minimum similarity score to return
	MinScore float32
}

// SearchResult represents a single search result
type SearchResult struct {
	// Document is the matched document
	Document *VectorDocument

	// Score is the similarity score
	// Interpretation depends on the distance metric:
	// - Cosine: -1 to 1 (higher is more similar)
	// - Euclidean: 0 to infinity (lower is more similar)
	// - DotProduct: -infinity to infinity (higher is more similar)
	Score float32

	// Rank is the position in the result set (1-based)
	Rank int
}

// VectorStoreError represents a vector store operation error
type VectorStoreError struct {
	Op         string // Operation that failed (e.g., "Add", "Search")
	Collection string // Collection name (if applicable)
	Err        error  // Underlying error
}

func (e *VectorStoreError) Error() string {
	if e.Collection != "" {
		return "vector store " + e.Op + " on collection '" + e.Collection + "': " + e.Err.Error()
	}
	return "vector store " + e.Op + ": " + e.Err.Error()
}

func (e *VectorStoreError) Unwrap() error {
	return e.Err
}

// NewVectorStoreError creates a new vector store error
func NewVectorStoreError(op, collection string, err error) error {
	return &VectorStoreError{
		Op:         op,
		Collection: collection,
		Err:        err,
	}
}

// DefaultSearchRequest creates a search request with sensible defaults
func DefaultSearchRequest(collection string, queryVector []float32) *SearchRequest {
	return &SearchRequest{
		Collection:       collection,
		QueryVector:      queryVector,
		TopK:             10,
		Filter:           nil,
		IncludeMetadata:  true,
		IncludeContent:   true,
		IncludeEmbedding: false,
		MinScore:         0.0,
	}
}

// DefaultTextSearchRequest creates a text search request with sensible defaults
func DefaultTextSearchRequest(collection string, query string) *TextSearchRequest {
	return &TextSearchRequest{
		Collection:       collection,
		Query:            query,
		TopK:             10,
		Filter:           nil,
		IncludeMetadata:  true,
		IncludeContent:   true,
		IncludeEmbedding: false,
		MinScore:         0.0,
	}
}
