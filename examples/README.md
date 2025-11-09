# Examples

This directory contains comprehensive examples demonstrating various features of the go-deep-agent library.

## Prerequisites

All examples require an OpenAI API key (except Ollama example):

```bash
export OPENAI_API_KEY="your-api-key-here"
```

## Available Examples

### 1. Basic Agent (ollama_example.go)

Demonstrates basic agent functionality with Ollama integration.

```bash
# Ensure Ollama is running
ollama serve

# Run example
go run examples/ollama_example.go
```

### 2. Batch Processing (batch_processing.go)

Shows how to process multiple prompts concurrently with various configurations.

**Features demonstrated:**
- Simple batch processing
- Progress tracking with callbacks
- Concurrency control
- Per-item completion callbacks
- Batch statistics (tokens, success rate)
- Retry mechanisms

```bash
go run examples/batch_processing.go
```

**Key Examples:**
- `simpleBatch()` - Process 5 prompts in parallel
- `batchWithProgress()` - Track progress with real-time updates
- `batchWithConcurrency()` - Control worker pool size (3 workers, 500ms delay)
- `batchWithCallbacks()` - Per-item completion notifications
- `batchStats()` - Aggregate metrics (tokens, success/failure rates)
- `batchWithRetry()` - Automatic retry up to 3 times

### 3. RAG (Retrieval-Augmented Generation) (rag_example.go)

Demonstrates RAG functionality for context-aware responses using document retrieval.

**Features demonstrated:**
- Basic RAG with document chunks
- Custom RAG configuration (chunk size, overlap, TopK)
- Document metadata management
- Custom retriever functions
- Retrieving and inspecting relevant documents
- Documentation Q&A system

```bash
go run examples/rag_example.go
```

**Key Examples:**
- `basicRAG()` - Simple knowledge base about programming languages
- `ragWithConfig()` - Custom chunk size (300), overlap (50), and TopK (1)
- `ragWithMetadata()` - Documents with source tracking
- `customRetriever()` - Custom lookup function for external knowledge bases
- `inspectRetrievedDocs()` - Examine retrieved documents and relevance scores
- `documentationQA()` - API documentation assistant with low temperature (0.3)

### 4. Multimodal (Vision) (builder_multimodal.go)

Image analysis with GPT-4 Vision.

**Features:**
- Describe images from URL
- Compare multiple images
- Control detail levels (low/high)
- OCR - extract text from images
- Analyze local image files
- Chart/graph analysis
- Multi-turn conversation with images

```bash
go run examples/builder_multimodal.go
```

### 5. Redis Cache (cache_redis_example.go)

Demonstrates Redis-based distributed caching for AI responses.

**Features demonstrated:**
- Simple Redis cache setup
- Advanced configuration with connection pooling
- Cache statistics tracking (hits, misses, hit rate)
- Batch operations for multiple queries
- Pattern-based cache deletion
- Distributed locking with SetNX
- Performance comparison (no cache vs memory cache vs Redis cache)
- TTL management (default, custom, disable/enable)

```bash
# Ensure Redis is running
redis-server

# Run example
go run examples/cache_redis_example.go
```

**Key Examples:**
- `example1SimpleRedisCache()` - Basic setup with localhost:6379, cache hit vs miss comparison
- `example2RedisCacheWithOptions()` - Custom pool size (20), key prefix ("myapp"), 10-minute TTL
- `example3CacheStatistics()` - Track hits, misses, writes, size, hit rate percentage
- `example4BatchOperations()` - Process 5 questions, compare uncached vs cached performance
- `example5PatternDeletion()` - Clear all cached entries at once
- `example6DistributedLocking()` - SetNX for cache stampede prevention, concurrent request handling
- `example7PerformanceComparison()` - Benchmark: no cache vs memory cache vs Redis cache
- `example8TTLManagement()` - Default TTL (5m), custom TTL (1h), disable/enable cache

**Performance Results:**
```
First call (cache miss): 1.2s
Second call (cache hit): 12ms
Speed improvement: 100x faster

Memory vs Redis (2nd call): 1.2x difference
Note: Memory cache is fastest but not shared across instances
Redis cache is slightly slower but shared and persistent
```

## Running Examples

Each example is self-contained and can be run independently:

```bash
# Run specific example
go run examples/batch_processing.go
go run examples/rag_example.go
go run examples/ollama_example.go
go run examples/cache_redis_example.go

# Or build and run
go build examples/batch_processing.go
./batch_processing
```

## Example Output

### Batch Processing Output:
```
Batch processing completed!
Total: 5 prompts
Success: 5, Failed: 0
Total tokens: 850 (prompt: 250, completion: 600)
```

### RAG Output:
```
Question: Who created the Go programming language?
Answer: Go was created by Robert Griesemer, Rob Pike, and Ken Thompson at Google.

Retrieved 2 documents:
  1. Source: golang.org (Score: 0.95)
  2. Source: go-history.md (Score: 0.78)
```

## Configuration Options

### Batch Processing Options:
- `MaxConcurrency` - Worker pool size (default: 5, range: 1-100)
- `DelayBetweenBatches` - Delay between request batches
- `ContinueOnError` - Continue processing if individual requests fail
- `OnProgress` - Callback for progress tracking (completed/total)
- `OnItemComplete` - Callback when each item completes

### RAG Options:
- `ChunkSize` - Size of document chunks (default: 1000 chars)
- `ChunkOverlap` - Overlap between chunks (default: 200 chars)
- `TopK` - Number of chunks to retrieve (default: 3)
- `MinScore` - Minimum relevance score threshold (default: 0.0)
- `Separator` - Separator between retrieved chunks
- `IncludeScores` - Show relevance scores in context

## Notes

- All examples use `gpt-4o-mini` model by default
- Batch processing respects rate limits with configurable delays
- RAG uses TF-IDF-based similarity scoring with exact phrase bonuses
- Custom retrievers enable integration with vector databases or external APIs
- Examples include error handling and graceful degradation

## Next Steps

After running these examples, explore:
- Combining batch processing with RAG for large-scale Q&A
- Implementing custom retrievers for your specific knowledge base
- Experimenting with different chunk sizes and TopK values
- Building production systems with token tracking and monitoring
