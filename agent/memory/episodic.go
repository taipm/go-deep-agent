package memory

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// VectorStoreAdapter is an interface for vector database operations
// This prevents import cycles with the agent package
type VectorStoreAdapter interface {
	// Add inserts documents with embeddings and metadata
	Add(ctx context.Context, collection string, docs []VectorDoc) ([]string, error)

	// SearchByText performs semantic search using text query
	SearchByText(ctx context.Context, req TextSearchReq) ([]SearchRes, error)

	// Delete removes documents by IDs
	Delete(ctx context.Context, collection string, ids []string) error

	// Count returns the number of documents in a collection
	Count(ctx context.Context, collection string) (int64, error)

	// Clear removes all documents from a collection
	Clear(ctx context.Context, collection string) error
}

// EmbeddingAdapter is an interface for generating text embeddings
type EmbeddingAdapter interface {
	// Embed generates an embedding vector for text
	Embed(ctx context.Context, text string) ([]float32, error)

	// Dimensions returns the dimensionality of embeddings
	Dimensions() int
}

// VectorDoc represents a document in the vector store
type VectorDoc struct {
	ID        string
	Content   string
	Embedding []float32
	Metadata  map[string]interface{}
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TextSearchReq contains parameters for text-based search
type TextSearchReq struct {
	Collection      string
	Query           string
	TopK            int
	Filter          map[string]interface{}
	IncludeMetadata bool
	IncludeContent  bool
	MinScore        float32
}

// SearchRes represents a search result
type SearchRes struct {
	Document VectorDoc
	Score    float32
	Rank     int
}

// EpisodicMemoryImpl implements the EpisodicMemory interface
// Vector-based implementation with semantic search capabilities
type EpisodicMemoryImpl struct {
	// Vector store for semantic search (optional)
	vectorStore VectorStoreAdapter
	embedding   EmbeddingAdapter
	collection  string

	// In-memory fallback storage (always used as cache)
	messages []MessageWithImportance
	mu       sync.RWMutex

	// Configuration
	useVectorStore bool
	maxSize        int
}

// MessageWithImportance wraps a message with its importance score
type MessageWithImportance struct {
	Message    Message
	Importance float64
}

// EpisodicMemoryConfig configures episodic memory behavior
type EpisodicMemoryConfig struct {
	VectorStore    VectorStoreAdapter // Optional vector store for semantic search
	Embedding      EmbeddingAdapter   // Embedding provider (required if using vector store)
	CollectionName string             // Collection name in vector store
	MaxSize        int                // Maximum messages to store (0 = unlimited)
}

// NewEpisodicMemory creates a new episodic memory with in-memory storage only
func NewEpisodicMemory() *EpisodicMemoryImpl {
	return &EpisodicMemoryImpl{
		messages:       make([]MessageWithImportance, 0),
		useVectorStore: false,
		maxSize:        0, // unlimited
	}
}

// NewEpisodicMemoryWithConfig creates episodic memory with custom configuration
func NewEpisodicMemoryWithConfig(config EpisodicMemoryConfig) *EpisodicMemoryImpl {
	em := &EpisodicMemoryImpl{
		messages:       make([]MessageWithImportance, 0),
		vectorStore:    config.VectorStore,
		embedding:      config.Embedding,
		collection:     config.CollectionName,
		maxSize:        config.MaxSize,
		useVectorStore: config.VectorStore != nil && config.Embedding != nil,
	}

	// Set default collection name if not provided
	if em.useVectorStore && em.collection == "" {
		em.collection = "episodic_memory"
	}

	return em
}

// Store implements EpisodicMemory.Store
func (e *EpisodicMemoryImpl) Store(ctx context.Context, msg Message, importance float64) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Check for duplicates
	if e.isDuplicate(msg) {
		return nil // Skip duplicate
	}

	// Add to in-memory storage
	e.messages = append(e.messages, MessageWithImportance{
		Message:    msg,
		Importance: importance,
	})

	// Enforce max size if configured
	if e.maxSize > 0 && len(e.messages) > e.maxSize {
		// Remove oldest messages
		e.messages = e.messages[len(e.messages)-e.maxSize:]
	}

	// Store in vector store if enabled
	if e.useVectorStore {
		if err := e.storeInVectorDB(ctx, msg, importance); err != nil {
			// Don't fail, just continue with in-memory storage
			// TODO: Add proper logging
			_ = err
		}
	}

	return nil
}

// storeInVectorDB stores a message in the vector database
func (e *EpisodicMemoryImpl) storeInVectorDB(ctx context.Context, msg Message, importance float64) error {
	// Prepare metadata
	metadata := make(map[string]interface{})
	metadata["role"] = msg.Role
	metadata["importance"] = importance
	metadata["timestamp"] = msg.Timestamp.Unix()

	// Copy original metadata
	for k, v := range msg.Metadata {
		metadata[k] = v
	}

	// Create vector document
	doc := VectorDoc{
		ID:        fmt.Sprintf("%d_%s", msg.Timestamp.UnixNano(), msg.Role),
		Content:   msg.Content,
		Metadata:  metadata,
		CreatedAt: msg.Timestamp,
		UpdatedAt: msg.Timestamp,
	}

	// Add to vector store
	_, err := e.vectorStore.Add(ctx, e.collection, []VectorDoc{doc})
	return err
}

// StoreBatch implements EpisodicMemory.StoreBatch
func (e *EpisodicMemoryImpl) StoreBatch(ctx context.Context, messages []Message, importances []float64) error {
	if len(messages) != len(importances) {
		return fmt.Errorf("messages and importances length mismatch")
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	// Filter out duplicates and collect valid messages
	validMessages := make([]Message, 0, len(messages))
	validImportances := make([]float64, 0, len(importances))

	for i, msg := range messages {
		if !e.isDuplicate(msg) {
			validMessages = append(validMessages, msg)
			validImportances = append(validImportances, importances[i])

			// Add to in-memory storage
			e.messages = append(e.messages, MessageWithImportance{
				Message:    msg,
				Importance: importances[i],
			})
		}
	}

	// Enforce max size if configured
	if e.maxSize > 0 && len(e.messages) > e.maxSize {
		e.messages = e.messages[len(e.messages)-e.maxSize:]
	}

	// Store in vector store if enabled and we have valid messages
	if e.useVectorStore && len(validMessages) > 0 {
		if err := e.storeBatchInVectorDB(ctx, validMessages, validImportances); err != nil {
			// Don't fail, continue with in-memory storage
			_ = err
		}
	}

	return nil
}

// storeBatchInVectorDB stores multiple messages in the vector database
func (e *EpisodicMemoryImpl) storeBatchInVectorDB(ctx context.Context, messages []Message, importances []float64) error {
	docs := make([]VectorDoc, len(messages))

	for i, msg := range messages {
		// Prepare metadata
		metadata := make(map[string]interface{})
		metadata["role"] = msg.Role
		metadata["importance"] = importances[i]
		metadata["timestamp"] = msg.Timestamp.Unix()

		// Copy original metadata
		for k, v := range msg.Metadata {
			metadata[k] = v
		}

		docs[i] = VectorDoc{
			ID:        fmt.Sprintf("%d_%s_%d", msg.Timestamp.UnixNano(), msg.Role, i),
			Content:   msg.Content,
			Metadata:  metadata,
			CreatedAt: msg.Timestamp,
			UpdatedAt: msg.Timestamp,
		}
	}

	// Add all documents to vector store
	_, err := e.vectorStore.Add(ctx, e.collection, docs)
	return err
}

// Retrieve implements EpisodicMemory.Retrieve
// Uses semantic similarity search with vector embeddings when available
func (e *EpisodicMemoryImpl) Retrieve(ctx context.Context, query string, topK int) ([]Message, error) {
	// Use vector store for semantic search if available
	if e.useVectorStore {
		return e.retrieveFromVectorDB(ctx, query, topK)
	}

	// Fallback to in-memory search (simple recency-based)
	e.mu.RLock()
	defer e.mu.RUnlock()

	size := len(e.messages)
	if topK > size {
		topK = size
	}

	result := make([]Message, 0, topK)
	for i := size - topK; i < size; i++ {
		result = append(result, e.messages[i].Message)
	}

	return result, nil
}

// retrieveFromVectorDB performs semantic search using vector store
func (e *EpisodicMemoryImpl) retrieveFromVectorDB(ctx context.Context, query string, topK int) ([]Message, error) {
	// Perform semantic search
	req := TextSearchReq{
		Collection:      e.collection,
		Query:           query,
		TopK:            topK,
		IncludeMetadata: true,
		IncludeContent:  true,
		MinScore:        0.0,
	}

	results, err := e.vectorStore.SearchByText(ctx, req)
	if err != nil {
		// Fallback to in-memory if vector search fails
		return e.retrieveInMemory(ctx, query, topK)
	}

	// Convert search results to messages
	messages := make([]Message, 0, len(results))
	for _, res := range results {
		msg := e.vectorDocToMessage(res.Document)
		messages = append(messages, msg)
	}

	return messages, nil
}

// retrieveInMemory performs simple in-memory retrieval
func (e *EpisodicMemoryImpl) retrieveInMemory(ctx context.Context, query string, topK int) ([]Message, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	size := len(e.messages)
	if topK > size {
		topK = size
	}

	result := make([]Message, 0, topK)
	for i := size - topK; i < size; i++ {
		result = append(result, e.messages[i].Message)
	}

	return result, nil
}

// vectorDocToMessage converts a VectorDoc to Message
func (e *EpisodicMemoryImpl) vectorDocToMessage(doc VectorDoc) Message {
	msg := Message{
		Content:   doc.Content,
		Timestamp: doc.CreatedAt,
		Metadata:  doc.Metadata,
	}

	// Extract role from metadata
	if role, ok := doc.Metadata["role"].(string); ok {
		msg.Role = role
	}

	return msg
}

// RetrieveByTime implements EpisodicMemory.RetrieveByTime
func (e *EpisodicMemoryImpl) RetrieveByTime(ctx context.Context, start, end time.Time, limit int) ([]Message, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result := make([]Message, 0)
	for _, m := range e.messages {
		if (m.Message.Timestamp.After(start) || m.Message.Timestamp.Equal(start)) &&
			(m.Message.Timestamp.Before(end) || m.Message.Timestamp.Equal(end)) {
			result = append(result, m.Message)
			if len(result) >= limit {
				break
			}
		}
	}

	return result, nil
}

// RetrieveByImportance implements EpisodicMemory.RetrieveByImportance
func (e *EpisodicMemoryImpl) RetrieveByImportance(ctx context.Context, minImportance float64, limit int) ([]Message, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result := make([]Message, 0)
	for _, m := range e.messages {
		if m.Importance >= minImportance {
			result = append(result, m.Message)
			if len(result) >= limit {
				break
			}
		}
	}

	return result, nil
}

// Search implements EpisodicMemory.Search
func (e *EpisodicMemoryImpl) Search(ctx context.Context, filter SearchFilter) ([]Message, error) {
	// Try vector store first if available and query is provided
	if e.useVectorStore && filter.Query != "" {
		return e.searchVectorDB(ctx, filter)
	}

	// Fallback to in-memory search
	return e.searchInMemory(ctx, filter)
}

// searchVectorDB performs search using vector store
func (e *EpisodicMemoryImpl) searchVectorDB(ctx context.Context, filter SearchFilter) ([]Message, error) {
	// Build filter metadata
	metadata := make(map[string]interface{})
	if filter.MinImportance > 0 {
		metadata["importance"] = map[string]interface{}{"$gte": filter.MinImportance}
	}
	if filter.TimeRange != nil {
		metadata["timestamp"] = map[string]interface{}{
			"$gte": filter.TimeRange.Start.Unix(),
			"$lte": filter.TimeRange.End.Unix(),
		}
	}

	req := TextSearchReq{
		Collection:      e.collection,
		Query:           filter.Query,
		TopK:            filter.Limit,
		Filter:          metadata,
		IncludeMetadata: true,
		IncludeContent:  true,
		MinScore:        0.0,
	}

	results, err := e.vectorStore.SearchByText(ctx, req)
	if err != nil {
		// Fallback to in-memory
		return e.searchInMemory(ctx, filter)
	}

	messages := make([]Message, 0, len(results))
	for _, res := range results {
		msg := e.vectorDocToMessage(res.Document)

		// Additional filtering for tags (vector store might not support this)
		if len(filter.Tags) > 0 && !hasTags(msg, filter.Tags) {
			continue
		}

		messages = append(messages, msg)
	}

	return messages, nil
}

// searchInMemory performs in-memory search with filters
func (e *EpisodicMemoryImpl) searchInMemory(ctx context.Context, filter SearchFilter) ([]Message, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result := make([]Message, 0)

	for _, m := range e.messages {
		if !e.matchesFilter(m, filter) {
			continue
		}

		result = append(result, m.Message)
		if len(result) >= filter.Limit {
			break
		}
	}

	return result, nil
}

// matchesFilter checks if a message matches the search filter
func (e *EpisodicMemoryImpl) matchesFilter(m MessageWithImportance, filter SearchFilter) bool {
	// Check importance
	if m.Importance < filter.MinImportance {
		return false
	}

	// Check time range
	if filter.TimeRange != nil {
		if m.Message.Timestamp.Before(filter.TimeRange.Start) ||
			m.Message.Timestamp.After(filter.TimeRange.End) {
			return false
		}
	}

	// Check tags
	if len(filter.Tags) > 0 && !hasTags(m.Message, filter.Tags) {
		return false
	}

	return true
}

// Clear implements EpisodicMemory.Clear
func (e *EpisodicMemoryImpl) Clear(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Clear in-memory storage
	e.messages = make([]MessageWithImportance, 0)

	// Clear vector store if enabled
	if e.useVectorStore {
		if err := e.vectorStore.Clear(ctx, e.collection); err != nil {
			return fmt.Errorf("failed to clear vector store: %w", err)
		}
	}

	return nil
}

// Size implements EpisodicMemory.Size
func (e *EpisodicMemoryImpl) Size() int {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return len(e.messages)
}

// hasTags checks if message has all required tags
func hasTags(msg Message, requiredTags []string) bool {
	if msg.Metadata == nil {
		return false
	}

	tags, ok := msg.Metadata["tags"].([]string)
	if !ok {
		return false
	}

	tagMap := make(map[string]bool)
	for _, tag := range tags {
		tagMap[tag] = true
	}

	for _, required := range requiredTags {
		if !tagMap[required] {
			return false
		}
	}

	return true
}

// isDuplicate checks if a message is a duplicate of an existing message
// Uses exact content matching and timestamp proximity (within 1 second)
func (e *EpisodicMemoryImpl) isDuplicate(msg Message) bool {
	// Check recent messages for duplicates
	// Only check last 100 messages for performance
	start := len(e.messages) - 100
	if start < 0 {
		start = 0
	}

	for i := start; i < len(e.messages); i++ {
		existing := e.messages[i].Message

		// Exact content match
		if existing.Content == msg.Content {
			// Check if timestamps are very close (within 1 second)
			timeDiff := msg.Timestamp.Sub(existing.Timestamp)
			if timeDiff < 0 {
				timeDiff = -timeDiff
			}

			if timeDiff < time.Second {
				return true
			}
		}
	}

	return false
}
