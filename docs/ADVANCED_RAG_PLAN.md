# Advanced RAG Implementation Plan

## Overview
Extend current RAG support with vector database integration for production-scale semantic search.

## Current State (v0.4.0)
- ✅ Basic document chunking with sentence boundaries
- ✅ TF-IDF similarity scoring with word overlap
- ✅ In-memory retrieval with TopK and MinScore
- ✅ Custom retriever function support
- ✅ Document metadata tracking

**Limitations:**
- No semantic understanding (only keyword matching)
- Limited to small document sets (in-memory)
- No persistence
- Basic similarity scoring

## Goals (v0.5.0+)

### Phase 1: Vector Database Integration (v0.5.0)
**Target:** Support 3+ vector databases with semantic embeddings

#### 1.1 Embedding Provider Interface
```go
type EmbeddingProvider interface {
    // Generate embeddings for text
    Embed(ctx context.Context, text string) ([]float32, error)
    
    // Batch embedding generation
    EmbedBatch(ctx context.Context, texts []string) ([][]float32, error)
    
    // Get embedding dimensions
    Dimensions() int
    
    // Get model name
    Model() string
}
```

**Implementations:**
- `OpenAIEmbedding` - text-embedding-3-small/large (OpenAI)
- `OllamaEmbedding` - nomic-embed-text (Local)
- `CustomEmbedding` - User-provided function

#### 1.2 Vector Store Interface
```go
type VectorStore interface {
    // Add documents with embeddings
    Add(ctx context.Context, docs []Document, embeddings [][]float32) error
    
    // Search by vector similarity
    Search(ctx context.Context, query []float32, topK int) ([]Document, error)
    
    // Delete documents
    Delete(ctx context.Context, ids []string) error
    
    // Update document metadata
    Update(ctx context.Context, id string, metadata map[string]string) error
    
    // Get statistics
    Stats() VectorStoreStats
}

type VectorStoreStats struct {
    TotalDocuments int
    TotalChunks    int
    IndexSize      int64
    LastUpdated    time.Time
}
```

**Implementations (Priority Order):**

1. **Chroma** (Priority 1 - Local, easy setup)
   - Go client: github.com/amikos-tech/chroma-go
   - Features: Local, Python-based, simple API
   - Use case: Development, small-scale production
   
2. **Qdrant** (Priority 2 - Production-ready)
   - Go client: github.com/qdrant/go-client
   - Features: Cloud + self-hosted, high performance
   - Use case: Production, scalable deployments
   
3. **Weaviate** (Priority 3 - Advanced features)
   - Go client: github.com/weaviate/weaviate-go-client
   - Features: GraphQL, hybrid search, multi-modal
   - Use case: Complex RAG pipelines
   
4. **Pinecone** (Priority 4 - Managed service)
   - REST API client
   - Features: Serverless, managed, scalable
   - Use case: Cloud-native applications

#### 1.3 Enhanced RAG Builder API
```go
// Vector-based RAG
builder.WithVectorRAG(
    agent.NewOpenAIEmbedding("text-embedding-3-small", apiKey),
    agent.NewChromaStore("http://localhost:8000"),
)

// Hybrid RAG (keyword + semantic)
builder.WithHybridRAG(
    agent.NewOpenAIEmbedding("text-embedding-3-small", apiKey),
    agent.NewQdrantStore("localhost:6333"),
    &agent.HybridRAGConfig{
        SemanticWeight: 0.7,  // 70% semantic, 30% keyword
        KeywordWeight:  0.3,
        RerankTopK:     10,   // Retrieve 10, then rerank to TopK
    },
)

// Advanced configuration
builder.WithRAGConfig(&agent.RAGConfig{
    ChunkSize:      1000,
    ChunkOverlap:   200,
    TopK:           5,
    MinScore:       0.7,
    UseReranking:   true,        // NEW: Rerank with cross-encoder
    UseHybrid:      true,         // NEW: Combine keyword + semantic
    CacheEmbeddings: true,        // NEW: Cache embeddings
    BatchSize:      100,          // NEW: Batch embedding generation
})
```

#### 1.4 Implementation Files

**agent/embedding.go** (~300 LOC)
- `EmbeddingProvider` interface
- `OpenAIEmbedding` implementation
- `OllamaEmbedding` implementation
- Batch processing
- Embedding caching

**agent/vectorstore.go** (~200 LOC)
- `VectorStore` interface
- `VectorStoreStats` struct
- Common utilities (cosine similarity, etc.)

**agent/vectorstore_chroma.go** (~400 LOC)
- Chroma client implementation
- Collection management
- CRUD operations
- Error handling

**agent/vectorstore_qdrant.go** (~400 LOC)
- Qdrant client implementation
- Collection management
- Hybrid search support

**agent/rag_vector.go** (~300 LOC)
- Vector-based retrieval
- Hybrid search (keyword + semantic)
- Reranking logic
- Integration with existing RAG

**Testing Strategy:**
- `embedding_test.go` - 15+ tests for embedding providers
- `vectorstore_test.go` - 20+ tests per store implementation
- `rag_vector_test.go` - 25+ tests for vector retrieval
- Integration tests with real vector DBs (Docker containers)

**Examples:**
- `examples/rag_vector_chroma.go` - Chroma integration
- `examples/rag_vector_qdrant.go` - Qdrant integration
- `examples/rag_hybrid.go` - Hybrid search
- `examples/rag_reranking.go` - Reranking demonstration

### Phase 2: Advanced Features (v0.6.0)

#### 2.1 Semantic Reranking
- Cross-encoder models for reranking
- Integration with sentence-transformers
- Configurable reranking strategies

#### 2.2 Multi-Query RAG
```go
builder.WithMultiQueryRAG(&agent.MultiQueryRAGConfig{
    NumQueries: 3,              // Generate 3 variations of query
    AggregationStrategy: "rrf", // Reciprocal Rank Fusion
})
```

#### 2.3 Contextual Compression
- Filter irrelevant chunks after retrieval
- Extract only relevant sentences
- Reduce context window usage

#### 2.4 Parent-Child Chunking
- Retrieve child chunks (small, specific)
- Return parent chunks (larger context)
- Better context preservation

## Implementation Timeline

### Sprint 1 (Week 1-2): Foundation
- [ ] Design and implement `EmbeddingProvider` interface
- [ ] Implement `OpenAIEmbedding`
- [ ] Implement `OllamaEmbedding`
- [ ] Create `VectorStore` interface
- [ ] Unit tests for embedding providers

### Sprint 2 (Week 3-4): Chroma Integration
- [ ] Implement `ChromaStore`
- [ ] CRUD operations
- [ ] Search functionality
- [ ] Integration tests with Chroma (Docker)
- [ ] Example: `rag_vector_chroma.go`

### Sprint 3 (Week 5-6): Qdrant Integration
- [ ] Implement `QdrantStore`
- [ ] Hybrid search support
- [ ] Performance optimization
- [ ] Integration tests with Qdrant
- [ ] Example: `rag_vector_qdrant.go`

### Sprint 4 (Week 7-8): Advanced Features
- [ ] Implement hybrid search (keyword + semantic)
- [ ] Add reranking logic
- [ ] Embedding caching
- [ ] Comprehensive documentation
- [ ] Performance benchmarks

## Success Metrics

### Functional
- ✅ Support 3+ vector databases
- ✅ Embedding generation (OpenAI, Ollama)
- ✅ Semantic search with cosine similarity
- ✅ Hybrid search (keyword + vector)
- ✅ 80+ tests, all passing
- ✅ 5+ working examples

### Performance
- Embedding generation: <100ms for single text
- Vector search: <50ms for 10K documents (Chroma)
- Vector search: <20ms for 100K documents (Qdrant)
- Batch embedding: 100 texts in <2s

### Developer Experience
- Simple API: 3-5 lines to enable vector RAG
- Clear examples for each vector store
- Comprehensive error messages
- Production-ready defaults

## Dependencies

### New Dependencies
```go
require (
    github.com/amikos-tech/chroma-go v0.1.0       // Chroma client
    github.com/qdrant/go-client v1.7.0            // Qdrant client
    github.com/weaviate/weaviate-go-client v4.13.1 // Optional: Weaviate
)
```

### Optional Dependencies
- Docker (for testing)
- Chroma server (local development)
- Qdrant server (local development)

## Migration Path

### For Existing Users (v0.4.0 → v0.5.0)
```go
// Old (still works)
builder.WithRAG(docs...)

// New (opt-in)
builder.WithVectorRAG(
    agent.NewOpenAIEmbedding("text-embedding-3-small", apiKey),
    agent.NewChromaStore("http://localhost:8000"),
)

// Both can coexist
builder.WithRAG(docs...).                    // Keyword-based (fast, simple)
    WithVectorRAG(embedder, vectorStore)     // Semantic (slower, better)
```

**No Breaking Changes** - Existing RAG functionality remains unchanged.

## Documentation Updates

### New Documentation Files
- `docs/RAG_VECTOR_DATABASES.md` - Vector DB comparison and setup
- `docs/RAG_EMBEDDINGS.md` - Embedding providers guide
- `docs/RAG_HYBRID_SEARCH.md` - Hybrid search strategies
- `docs/RAG_PERFORMANCE.md` - Performance tuning guide

### Updated Documentation
- `README.md` - Add vector RAG examples
- `ROADMAP.md` - Update with v0.5.0 details
- `examples/README.md` - Add vector RAG examples

## Risk Mitigation

### Technical Risks
1. **Vector DB availability** - Provide fallback to in-memory
2. **Embedding API costs** - Support local models (Ollama)
3. **Performance** - Batch operations, caching
4. **Compatibility** - Extensive integration testing

### User Impact
- Backward compatible
- Clear migration guide
- Optional features
- Good defaults

## Future Considerations (v0.7.0+)

### Advanced Vector DB Features
- Weaviate integration (GraphQL, multi-modal)
- Pinecone integration (serverless)
- Milvus integration (large-scale)

### Advanced RAG Patterns
- Self-querying retrieval
- Iterative retrieval
- Retrieval with feedback
- Multi-hop reasoning

### Enterprise Features
- Vector store sharding
- Distributed embedding generation
- Advanced caching strategies
- Monitoring and observability
