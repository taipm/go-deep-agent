package agent

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
)

// Document represents a document chunk for RAG
type Document struct {
	Content  string            // Document text content
	Metadata map[string]string // Optional metadata (source, page, etc.)
	Score    float64           // Relevance score (set during retrieval)
}

// RAGConfig configures RAG behavior
type RAGConfig struct {
	// ChunkSize is the maximum size of each document chunk (default: 1000)
	ChunkSize int

	// ChunkOverlap is the number of characters to overlap between chunks (default: 200)
	ChunkOverlap int

	// TopK is the number of most relevant documents to retrieve (default: 3)
	TopK int

	// MinScore is the minimum relevance score to include (0.0-1.0, default: 0.0)
	MinScore float64

	// Separator is the text used to join retrieved documents (default: "\n\n---\n\n")
	Separator string

	// IncludeScores adds relevance scores to the context (default: false)
	IncludeScores bool
}

// DefaultRAGConfig returns default RAG configuration
func DefaultRAGConfig() *RAGConfig {
	return &RAGConfig{
		ChunkSize:     1000,
		ChunkOverlap:  200,
		TopK:          3,
		MinScore:      0.0,
		Separator:     "\n\n---\n\n",
		IncludeScores: false,
	}
}

// RAGRetriever is a function that retrieves relevant documents for a query
type RAGRetriever func(query string) ([]Document, error)

// WithRAG enables RAG with provided documents
// Documents are automatically chunked and retrieved based on relevance
func (b *Builder) WithRAG(documents ...string) *Builder {
	if len(documents) == 0 {
		return b
	}

	// Convert strings to Document objects
	docs := make([]Document, len(documents))
	for i, doc := range documents {
		docs[i] = Document{
			Content:  doc,
			Metadata: map[string]string{"index": fmt.Sprintf("%d", i)},
		}
	}

	b.ragDocuments = docs
	b.ragEnabled = true

	if b.ragConfig == nil {
		b.ragConfig = DefaultRAGConfig()
	}

	return b
}

// WithRAGDocuments enables RAG with Document objects (with metadata)
func (b *Builder) WithRAGDocuments(documents ...Document) *Builder {
	if len(documents) == 0 {
		return b
	}

	b.ragDocuments = documents
	b.ragEnabled = true

	if b.ragConfig == nil {
		b.ragConfig = DefaultRAGConfig()
	}

	return b
}

// WithRAGRetriever sets a custom retriever function
func (b *Builder) WithRAGRetriever(retriever RAGRetriever) *Builder {
	b.ragRetriever = retriever
	b.ragEnabled = true

	if b.ragConfig == nil {
		b.ragConfig = DefaultRAGConfig()
	}

	return b
}

// WithRAGConfig sets custom RAG configuration
func (b *Builder) WithRAGConfig(config *RAGConfig) *Builder {
	b.ragConfig = config
	return b
}

// WithRAGTopK sets the number of documents to retrieve
func (b *Builder) WithRAGTopK(k int) *Builder {
	if k <= 0 {
		k = 3
	}

	if b.ragConfig == nil {
		b.ragConfig = DefaultRAGConfig()
	}

	b.ragConfig.TopK = k
	return b
}

// WithRAGChunkSize sets the document chunk size
func (b *Builder) WithRAGChunkSize(size int) *Builder {
	if size <= 0 {
		size = 1000
	}

	if b.ragConfig == nil {
		b.ragConfig = DefaultRAGConfig()
	}

	b.ragConfig.ChunkSize = size
	return b
}

// ChunkDocument splits a document into smaller chunks with overlap
func ChunkDocument(text string, chunkSize, overlap int) []string {
	if chunkSize <= 0 {
		chunkSize = 1000
	}
	if overlap < 0 {
		overlap = 0
	}
	if overlap >= chunkSize {
		overlap = chunkSize / 2
	}

	// Normalize whitespace
	text = strings.TrimSpace(text)
	if text == "" {
		return []string{}
	}

	var chunks []string
	textLen := len(text)
	start := 0

	for start < textLen {
		// Calculate end position
		end := start + chunkSize
		if end >= textLen {
			// Last chunk - take everything remaining
			chunk := strings.TrimSpace(text[start:])
			if chunk != "" {
				chunks = append(chunks, chunk)
			}
			break
		}

		// Try to break at sentence boundary
		breakPoints := []string{". ", ".\n", "! ", "!\n", "? ", "?\n"}
		bestBreak := -1
		searchStart := start + chunkSize/2 // Don't look too early

		for i := end; i >= searchStart; i-- {
			for _, bp := range breakPoints {
				if i+len(bp) <= textLen && text[i:i+len(bp)] == bp {
					bestBreak = i + len(bp)
					break
				}
			}
			if bestBreak > 0 {
				break
			}
		}

		if bestBreak > 0 {
			end = bestBreak
		}

		chunk := strings.TrimSpace(text[start:end])
		if chunk != "" {
			chunks = append(chunks, chunk)
		}

		// Move to next chunk with overlap
		start = end - overlap
	}

	return chunks
}

// retrieveRelevantDocs retrieves the most relevant documents for a query
func (b *Builder) retrieveRelevantDocs(ctx context.Context, query string) ([]Document, error) {
	logger := b.getLogger()
	logger.Debug(ctx, "RAG retrieval started", F("query_length", len(query)))

	// If vector store is configured, use vector search
	if b.vectorStore != nil && b.embeddingProvider != nil {
		logger.Debug(ctx, "Using vector store for retrieval",
			F("provider", fmt.Sprintf("%T", b.embeddingProvider)),
			F("store", fmt.Sprintf("%T", b.vectorStore)))
		return b.retrieveFromVector(ctx, query)
	}

	// If custom retriever is set, use it
	if b.ragRetriever != nil {
		logger.Debug(ctx, "Using custom retriever")
		docs, err := b.ragRetriever(query)
		if err != nil {
			logger.Error(ctx, "Custom retriever failed", F("error", err.Error()))
			return nil, err
		}
		logger.Debug(ctx, "Custom retriever completed", F("doc_count", len(docs)))
		return docs, nil
	}

	// If no documents, return empty
	if len(b.ragDocuments) == 0 {
		logger.Debug(ctx, "No RAG documents available")
		return []Document{}, nil
	}

	logger.Debug(ctx, "Using TF-IDF fallback retrieval", F("total_docs", len(b.ragDocuments)))

	config := b.ragConfig
	if config == nil {
		config = DefaultRAGConfig()
	}

	// Chunk all documents (TF-IDF fallback)
	var allChunks []Document
	for _, doc := range b.ragDocuments {
		chunks := ChunkDocument(doc.Content, config.ChunkSize, config.ChunkOverlap)
		for i, chunk := range chunks {
			chunkDoc := Document{
				Content:  chunk,
				Metadata: doc.Metadata,
			}
			// Add chunk index to metadata
			if chunkDoc.Metadata == nil {
				chunkDoc.Metadata = make(map[string]string)
			}
			chunkDoc.Metadata["chunk"] = fmt.Sprintf("%d", i)
			allChunks = append(allChunks, chunkDoc)
		}
	}

	logger.Debug(ctx, "Documents chunked",
		F("total_chunks", len(allChunks)),
		F("chunk_size", config.ChunkSize),
		F("chunk_overlap", config.ChunkOverlap))

	// Calculate relevance scores using simple similarity
	for i := range allChunks {
		allChunks[i].Score = calculateSimilarity(query, allChunks[i].Content)
	}

	// Sort by score (descending)
	sort.Slice(allChunks, func(i, j int) bool {
		return allChunks[i].Score > allChunks[j].Score
	})

	// Filter by minimum score and take top K
	var results []Document
	for i := 0; i < len(allChunks) && i < config.TopK; i++ {
		if allChunks[i].Score >= config.MinScore {
			results = append(results, allChunks[i])
		}
	}

	logger.Info(ctx, "RAG retrieval completed",
		F("results", len(results)),
		F("top_k", config.TopK),
		F("min_score", config.MinScore))

	return results, nil
}

// buildRAGContext builds the context string from retrieved documents
func (b *Builder) buildRAGContext(docs []Document) string {
	if len(docs) == 0 {
		return ""
	}

	config := b.ragConfig
	if config == nil {
		config = DefaultRAGConfig()
	}

	var parts []string
	for i, doc := range docs {
		if config.IncludeScores {
			parts = append(parts, fmt.Sprintf("[Document %d] (Relevance: %.2f)\n%s",
				i+1, doc.Score, doc.Content))
		} else {
			parts = append(parts, doc.Content)
		}
	}

	return strings.Join(parts, config.Separator)
}

// calculateSimilarity calculates similarity between query and document
// Uses simple word overlap + TF-IDF-like scoring
func calculateSimilarity(query, document string) float64 {
	queryWords := tokenize(strings.ToLower(query))
	docWords := tokenize(strings.ToLower(document))

	if len(queryWords) == 0 || len(docWords) == 0 {
		return 0.0
	}

	// Count word frequencies
	queryFreq := make(map[string]int)
	docFreq := make(map[string]int)

	for _, word := range queryWords {
		queryFreq[word]++
	}

	for _, word := range docWords {
		docFreq[word]++
	}

	// Calculate overlap score
	overlap := 0.0
	for word, qFreq := range queryFreq {
		if dFreq, exists := docFreq[word]; exists {
			// Weight by inverse frequency (rare words matter more)
			weight := 1.0 / math.Log(float64(dFreq+2))
			overlap += float64(min(qFreq, dFreq)) * weight
		}
	}

	// Normalize by query length
	score := overlap / float64(len(queryWords))

	// Bonus for exact phrase matches
	queryLower := strings.ToLower(query)
	docLower := strings.ToLower(document)
	if strings.Contains(docLower, queryLower) {
		score += 0.5
	}

	// Cap at 1.0
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// tokenize splits text into words
func tokenize(text string) []string {
	// Remove punctuation and split on whitespace
	reg := regexp.MustCompile(`[^\w\s]`)
	text = reg.ReplaceAllString(text, " ")

	words := strings.Fields(text)

	// Filter out very short words and common stop words
	stopWords := map[string]bool{
		"a": true, "an": true, "the": true, "and": true, "or": true,
		"but": true, "in": true, "on": true, "at": true, "to": true,
		"for": true, "of": true, "as": true, "by": true, "is": true,
		"are": true, "was": true, "were": true, "be": true, "been": true,
	}

	var filtered []string
	for _, word := range words {
		if len(word) > 2 && !stopWords[word] {
			filtered = append(filtered, word)
		}
	}

	return filtered
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GetLastRetrievedDocs returns the documents retrieved in the last RAG query
func (b *Builder) GetLastRetrievedDocs() []Document {
	return b.lastRetrievedDocs
}

// ClearRAG disables RAG and clears all RAG data
func (b *Builder) ClearRAG() *Builder {
	b.ragEnabled = false
	b.ragDocuments = nil
	b.ragRetriever = nil
	b.ragConfig = nil
	b.lastRetrievedDocs = nil
	b.vectorStore = nil
	b.embeddingProvider = nil
	b.vectorCollection = ""
	return b
}

// WithVectorRAG enables vector-based RAG using a vector database
// This provides semantic search capabilities for retrieval
func (b *Builder) WithVectorRAG(embedding EmbeddingProvider, store VectorStore, collection string) *Builder {
	b.embeddingProvider = embedding
	b.vectorStore = store
	b.vectorCollection = collection
	b.ragEnabled = true

	if b.ragConfig == nil {
		b.ragConfig = DefaultRAGConfig()
	}

	return b
}

// AddDocumentsToVector adds documents to the vector store
// Documents are automatically embedded and stored
func (b *Builder) AddDocumentsToVector(ctx context.Context, documents ...string) ([]string, error) {
	if b.vectorStore == nil || b.embeddingProvider == nil {
		return nil, fmt.Errorf("vector store and embedding provider must be configured with WithVectorRAG")
	}

	// Convert strings to VectorDocuments
	vectorDocs := make([]*VectorDocument, len(documents))
	for i, doc := range documents {
		vectorDocs[i] = &VectorDocument{
			Content: doc,
			Metadata: map[string]interface{}{
				"index": i,
			},
		}
	}

	// Add to vector store (embeddings will be auto-generated)
	return b.vectorStore.Add(ctx, b.vectorCollection, vectorDocs)
}

// AddVectorDocuments adds VectorDocument objects to the vector store
func (b *Builder) AddVectorDocuments(ctx context.Context, documents ...*VectorDocument) ([]string, error) {
	if b.vectorStore == nil {
		return nil, fmt.Errorf("vector store must be configured with WithVectorRAG")
	}

	return b.vectorStore.Add(ctx, b.vectorCollection, documents)
}

// retrieveFromVector performs semantic search using vector database
func (b *Builder) retrieveFromVector(ctx context.Context, query string) ([]Document, error) {
	logger := b.getLogger()
	logger.Debug(ctx, "Vector search started",
		F("collection", b.vectorCollection),
		F("query_length", len(query)))

	if b.vectorStore == nil || b.embeddingProvider == nil {
		err := fmt.Errorf("vector store and embedding provider must be configured")
		logger.Error(ctx, "Vector search configuration missing", F("error", err.Error()))
		return nil, err
	}

	config := b.ragConfig
	if config == nil {
		config = DefaultRAGConfig()
	}

	// Perform semantic search
	searchReq := &TextSearchRequest{
		Collection:      b.vectorCollection,
		Query:           query,
		TopK:            config.TopK,
		MinScore:        float32(config.MinScore),
		IncludeContent:  true,
		IncludeMetadata: true,
	}

	logger.Debug(ctx, "Executing vector search",
		F("top_k", config.TopK),
		F("min_score", config.MinScore))

	results, err := b.vectorStore.SearchByText(ctx, searchReq)
	if err != nil {
		logger.Error(ctx, "Vector search failed", F("error", err.Error()))
		return nil, fmt.Errorf("vector search failed: %w", err)
	}

	logger.Info(ctx, "Vector search completed",
		F("results", len(results)),
		F("collection", b.vectorCollection))

	// Convert SearchResults to Documents
	docs := make([]Document, len(results))
	for i, result := range results {
		// Convert metadata map[string]interface{} to map[string]string
		metadata := make(map[string]string)
		if result.Document.Metadata != nil {
			for k, v := range result.Document.Metadata {
				metadata[k] = fmt.Sprintf("%v", v)
			}
		}

		docs[i] = Document{
			Content:  result.Document.Content,
			Metadata: metadata,
			Score:    float64(result.Score),
		}
	}

	return docs, nil
}
