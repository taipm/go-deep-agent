package main

import (
	"context"
	"fmt"
	"log"

	agent "github.com/taipm/go-deep-agent/agent"
)

func main() {
	fmt.Println("=== Qdrant Vector Database Examples ===\n")

	// Prerequisites:
	// 1. Start Qdrant: docker run -p 6333:6333 qdrant/qdrant
	// 2. Start Ollama: ollama serve
	// 3. Pull model: ollama pull nomic-embed-text
	//
	// Alternative: Use OpenAI embeddings
	// embedding := agent.NewOpenAIEmbedding("YOUR_API_KEY", "text-embedding-3-small", 1536)

	ctx := context.Background()

	// Example 1: Setup Qdrant with Ollama embeddings
	fmt.Println("Example 1: Setup Qdrant Store")
	fmt.Println("------------------------------")

	// Create embedding provider using Ollama
	embedding, err := agent.NewOllamaEmbedding(
		"http://localhost:11434",
		"nomic-embed-text", // 768 dimensions
	)
	if err != nil {
		log.Fatalf("Failed to create embedding provider: %v", err)
	}

	// Create Qdrant store
	store, err := agent.NewQdrantStore("http://localhost:6333")
	if err != nil {
		log.Fatalf("Failed to create Qdrant store: %v", err)
	}

	// Configure with embedding provider for auto-embedding
	store.WithEmbedding(embedding)

	fmt.Println("✓ Qdrant store created and configured")
	fmt.Println()

	// Example 2: Create a collection
	fmt.Println("Example 2: Create Collection")
	fmt.Println("-----------------------------")

	collectionConfig := &agent.CollectionConfig{
		Name:           "go-docs",
		Description:    "Go programming documentation",
		Dimension:      768, // Must match embedding dimension
		DistanceMetric: agent.DistanceMetricCosine,
	}

	err = store.CreateCollection(ctx, "go-docs", collectionConfig)
	if err != nil {
		log.Printf("Warning: Collection might already exist: %v", err)
	} else {
		fmt.Println("✓ Collection 'go-docs' created")
	}
	fmt.Println()

	// Example 3: Add documents with auto-embedding
	fmt.Println("Example 3: Add Documents")
	fmt.Println("-------------------------")

	docs := []*agent.VectorDocument{
		{
			ID:      "go-intro",
			Content: "Go is a statically typed, compiled programming language designed at Google. It is syntactically similar to C, but with memory safety, garbage collection, structural typing, and CSP-style concurrency.",
			Metadata: map[string]interface{}{
				"category": "introduction",
				"level":    "beginner",
			},
		},
		{
			ID:      "go-goroutines",
			Content: "Goroutines are lightweight threads managed by the Go runtime. They are created using the 'go' keyword and enable concurrent execution of functions.",
			Metadata: map[string]interface{}{
				"category": "concurrency",
				"level":    "intermediate",
			},
		},
		{
			ID:      "go-channels",
			Content: "Channels are typed conduits through which you can send and receive values with the channel operator <-. They provide a way for goroutines to communicate and synchronize.",
			Metadata: map[string]interface{}{
				"category": "concurrency",
				"level":    "intermediate",
			},
		},
		{
			ID:      "go-interfaces",
			Content: "An interface type is defined as a set of method signatures. A value of interface type can hold any value that implements those methods. Interfaces enable polymorphism in Go.",
			Metadata: map[string]interface{}{
				"category": "types",
				"level":    "intermediate",
			},
		},
		{
			ID:      "go-packages",
			Content: "Packages are Go's way of organizing and reusing code. Every Go program is made up of packages. Programs start running in package main.",
			Metadata: map[string]interface{}{
				"category": "organization",
				"level":    "beginner",
			},
		},
	}

	// Add documents (embeddings will be auto-generated)
	ids, err := store.Add(ctx, "go-docs", docs)
	if err != nil {
		log.Fatalf("Failed to add documents: %v", err)
	}

	fmt.Printf("✓ Added %d documents: %v\n", len(ids), ids)
	fmt.Println()

	// Example 4: Semantic search by text
	fmt.Println("Example 4: Semantic Search")
	fmt.Println("---------------------------")

	searchReq := &agent.TextSearchRequest{
		Collection:      "go-docs",
		Query:           "How does concurrent programming work in Go?",
		TopK:            3,
		IncludeContent:  true,
		IncludeMetadata: true,
	}

	results, err := store.SearchByText(ctx, searchReq)
	if err != nil {
		log.Fatalf("Failed to search: %v", err)
	}

	fmt.Printf("Query: \"%s\"\n", searchReq.Query)
	fmt.Printf("Found %d results:\n\n", len(results))

	for _, result := range results {
		fmt.Printf("  %d. [Score: %.4f] %s\n", result.Rank, result.Score, result.Document.ID)
		fmt.Printf("     %s\n", result.Document.Content)
		fmt.Printf("     Category: %v\n", result.Document.Metadata["category"])
		fmt.Println()
	}

	// Example 5: Vector search with pre-computed embedding
	fmt.Println("Example 5: Vector Search")
	fmt.Println("------------------------")

	// Generate embedding for query
	queryEmb, err := embedding.Embed(ctx, "What are Go packages?")
	if err != nil {
		log.Fatalf("Failed to generate embedding: %v", err)
	}

	vectorSearchReq := &agent.SearchRequest{
		Collection:      "go-docs",
		QueryVector:     queryEmb,
		TopK:            2,
		IncludeContent:  true,
		IncludeMetadata: true,
	}

	results, err = store.Search(ctx, vectorSearchReq)
	if err != nil {
		log.Fatalf("Failed to vector search: %v", err)
	}

	fmt.Println("Query: \"What are Go packages?\"")
	fmt.Printf("Found %d results:\n\n", len(results))

	for _, result := range results {
		fmt.Printf("  %d. [Score: %.4f] %s\n", result.Rank, result.Score, result.Document.ID)
		fmt.Printf("     %s\n", result.Document.Content[:80]+"...")
		fmt.Println()
	}

	// Example 6: Filtered search by metadata
	fmt.Println("Example 6: Filtered Search")
	fmt.Println("--------------------------")

	filteredReq := &agent.TextSearchRequest{
		Collection: "go-docs",
		Query:      "Go programming concepts",
		TopK:       10,
		Filter: map[string]interface{}{
			"category": "concurrency",
		},
		IncludeContent:  true,
		IncludeMetadata: true,
	}

	results, err = store.SearchByText(ctx, filteredReq)
	if err != nil {
		log.Fatalf("Failed to filtered search: %v", err)
	}

	fmt.Printf("Query: \"%s\" with filter: category=concurrency\n", filteredReq.Query)
	fmt.Printf("Found %d results:\n\n", len(results))

	for _, result := range results {
		fmt.Printf("  %d. [Score: %.4f] %s\n", result.Rank, result.Score, result.Document.ID)
		fmt.Printf("     Category: %v, Level: %v\n", 
			result.Document.Metadata["category"], 
			result.Document.Metadata["level"])
		fmt.Println()
	}

	// Example 7: Get specific documents by ID
	fmt.Println("Example 7: Get Documents by ID")
	fmt.Println("-------------------------------")

	docIDs := []string{"go-intro", "go-channels"}
	retrievedDocs, err := store.Get(ctx, "go-docs", docIDs)
	if err != nil {
		log.Fatalf("Failed to get documents: %v", err)
	}

	fmt.Printf("Retrieved %d documents:\n\n", len(retrievedDocs))
	for _, doc := range retrievedDocs {
		fmt.Printf("  • %s: %s\n", doc.ID, doc.Content[:60]+"...")
	}
	fmt.Println()

	// Example 8: Update documents
	fmt.Println("Example 8: Update Document")
	fmt.Println("--------------------------")

	updateDocs := []*agent.VectorDocument{
		{
			ID:      "go-intro",
			Content: "Go (also known as Golang) is a statically typed, compiled programming language designed at Google by Robert Griesemer, Rob Pike, and Ken Thompson. It is syntactically similar to C but with memory safety, garbage collection, structural typing, and CSP-style concurrency. Go is widely used for cloud services, DevOps tools, and microservices.",
			Metadata: map[string]interface{}{
				"category": "introduction",
				"level":    "beginner",
				"updated":  true,
			},
		},
	}

	err = store.Update(ctx, "go-docs", updateDocs)
	if err != nil {
		log.Fatalf("Failed to update document: %v", err)
	}

	fmt.Println("✓ Updated document 'go-intro'")
	fmt.Println()

	// Example 9: Delete documents
	fmt.Println("Example 9: Delete Document")
	fmt.Println("--------------------------")

	countBefore, _ := store.Count(ctx, "go-docs")
	fmt.Printf("Documents before delete: %d\n", countBefore)

	err = store.Delete(ctx, "go-docs", []string{"go-packages"})
	if err != nil {
		log.Fatalf("Failed to delete document: %v", err)
	}

	countAfter, _ := store.Count(ctx, "go-docs")
	fmt.Printf("Documents after delete: %d\n", countAfter)
	fmt.Println("✓ Deleted document 'go-packages'")
	fmt.Println()

	// Example 10: Semantic Q&A system
	fmt.Println("Example 10: Semantic Q&A System")
	fmt.Println("--------------------------------")

	questions := []string{
		"What is Go?",
		"How do goroutines work?",
		"What are interfaces in Go?",
	}

	for i, question := range questions {
		fmt.Printf("\nQ%d: %s\n", i+1, question)

		qaReq := &agent.TextSearchRequest{
			Collection:     "go-docs",
			Query:          question,
			TopK:           1,
			IncludeContent: true,
		}

		qaResults, err := store.SearchByText(ctx, qaReq)
		if err != nil {
			log.Printf("Search failed: %v", err)
			continue
		}

		if len(qaResults) > 0 {
			answer := qaResults[0]
			fmt.Printf("A%d: %s\n", i+1, answer.Document.Content)
			fmt.Printf("    (Confidence: %.2f%%)\n", answer.Score*100)
		}
	}
	fmt.Println()

	// Example 11: Batch operations
	fmt.Println("Example 11: Batch Add Documents")
	fmt.Println("--------------------------------")

	batchDocs := []*agent.VectorDocument{
		{
			ID:      "go-testing",
			Content: "Go has a built-in testing framework provided by the 'testing' package. Tests are written in files ending with _test.go and test functions must start with Test.",
			Metadata: map[string]interface{}{
				"category": "testing",
				"level":    "intermediate",
			},
		},
		{
			ID:      "go-errors",
			Content: "Go uses explicit error handling. Functions that can fail return an error value as their last return value. The idiomatic way is to check if err != nil.",
			Metadata: map[string]interface{}{
				"category": "error-handling",
				"level":    "beginner",
			},
		},
		{
			ID:      "go-modules",
			Content: "Go modules are the official dependency management system. A module is a collection of packages with go.mod at its root. Use 'go mod init' to create a new module.",
			Metadata: map[string]interface{}{
				"category": "organization",
				"level":    "intermediate",
			},
		},
	}

	batchIDs, err := store.Add(ctx, "go-docs", batchDocs)
	if err != nil {
		log.Fatalf("Failed to batch add: %v", err)
	}

	fmt.Printf("✓ Batch added %d documents\n", len(batchIDs))

	totalCount, _ := store.Count(ctx, "go-docs")
	fmt.Printf("Total documents in collection: %d\n", totalCount)
	fmt.Println()

	// Example 12: List collections
	fmt.Println("Example 12: List All Collections")
	fmt.Println("---------------------------------")

	collections, err := store.ListCollections(ctx)
	if err != nil {
		log.Fatalf("Failed to list collections: %v", err)
	}

	fmt.Printf("Found %d collection(s):\n", len(collections))
	for _, name := range collections {
		count, _ := store.Count(ctx, name)
		fmt.Printf("  • %s (%d documents)\n", name, count)
	}
	fmt.Println()

	// Example 13: Score threshold search
	fmt.Println("Example 13: Search with Score Threshold")
	fmt.Println("----------------------------------------")

	thresholdReq := &agent.TextSearchRequest{
		Collection:     "go-docs",
		Query:          "concurrency patterns",
		TopK:           10,
		MinScore:       0.7, // Only return results with score >= 0.7
		IncludeContent: true,
	}

	results, err = store.SearchByText(ctx, thresholdReq)
	if err != nil {
		log.Fatalf("Failed to search with threshold: %v", err)
	}

	fmt.Printf("Query: \"%s\" (min score: %.1f)\n", thresholdReq.Query, thresholdReq.MinScore)
	fmt.Printf("Found %d high-confidence results:\n\n", len(results))

	for _, result := range results {
		fmt.Printf("  %d. [Score: %.4f] %s\n", result.Rank, result.Score, result.Document.ID)
	}
	fmt.Println()

	// Cleanup (uncomment to actually delete)
	fmt.Println("=== Cleanup ===")
	// err = store.Clear(ctx, "go-docs")
	// if err != nil {
	// 	log.Printf("Failed to clear collection: %v", err)
	// }
	// fmt.Println("✓ Collection cleared")

	// err = store.DeleteCollection(ctx, "go-docs")
	// if err != nil {
	// 	log.Printf("Failed to delete collection: %v", err)
	// }
	// fmt.Println("✓ Collection deleted")

	fmt.Println("\n✓ Examples completed successfully!")
	fmt.Println("\nNote: Qdrant offers advanced features:")
	fmt.Println("  - Point versioning and optimistic concurrency")
	fmt.Println("  - Payload indexing for faster filtering")
	fmt.Println("  - Advanced filtering with must/should/must_not")
	fmt.Println("  - Quantization for reduced memory usage")
	fmt.Println("  - Distributed deployment with sharding")
	fmt.Println("  - Snapshots and backups")
}
