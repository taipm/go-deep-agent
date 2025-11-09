package main

import (
	"context"
	"fmt"
	"log"

	"github.com/taipm/go-deep-agent/agent"
)

func main() {
	ctx := context.Background()

	// Example 1: Setup ChromaDB with Ollama Embeddings
	fmt.Println("=== Example 1: Setup ChromaDB with Ollama Embeddings ===")

	// Create Ollama embedding provider (local, free)
	embedder, err := agent.NewOllamaEmbedding(
		agent.OllamaEmbeddingModelNomic, // nomic-embed-text (768 dimensions)
		"http://localhost:11434",
	)
	if err != nil {
		log.Fatal(err)
	}

	// Create ChromaDB client
	store, err := agent.NewChromaStore("http://localhost:8000")
	if err != nil {
		log.Fatal(err)
	}

	// Configure with embedding provider
	store.WithEmbedding(embedder)

	fmt.Println("✓ ChromaDB client initialized")
	fmt.Printf("✓ Embedding provider: %s (%dd)\n\n", embedder.Model(), embedder.Dimensions())

	// Example 2: Create Collection
	fmt.Println("=== Example 2: Create Collection ===")

	collectionName := "go-docs"
	config := &agent.CollectionConfig{
		Name:           collectionName,
		Description:    "Go programming documentation",
		Dimension:      embedder.Dimensions(),
		DistanceMetric: agent.DistanceMetricCosine,
	}

	err = store.CreateCollection(ctx, collectionName, config)
	if err != nil {
		fmt.Printf("Note: Collection may already exist - %v\n", err)
	} else {
		fmt.Printf("✓ Collection '%s' created\n", collectionName)
	}

	// List collections
	collections, err := store.ListCollections(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ Available collections: %v\n\n", collections)

	// Example 3: Add Documents with Auto-Embedding
	fmt.Println("=== Example 3: Add Documents with Auto-Embedding ===")

	docs := []*agent.VectorDocument{
		{
			ID:      "go-intro",
			Content: "Go is a statically typed, compiled programming language designed at Google.",
			Metadata: map[string]interface{}{
				"category": "introduction",
				"source":   "golang.org",
			},
		},
		{
			ID:      "go-concurrency",
			Content: "Goroutines are lightweight threads managed by the Go runtime. They enable easy concurrent programming.",
			Metadata: map[string]interface{}{
				"category": "concurrency",
				"source":   "golang.org",
			},
		},
		{
			ID:      "go-interfaces",
			Content: "Interfaces in Go provide a way to specify the behavior of an object. A type implements an interface by implementing its methods.",
			Metadata: map[string]interface{}{
				"category": "types",
				"source":   "golang.org",
			},
		},
		{
			ID:      "go-packages",
			Content: "Go programs are organized into packages. A package is a collection of source files in the same directory.",
			Metadata: map[string]interface{}{
				"category": "structure",
				"source":   "golang.org",
			},
		},
		{
			ID:      "go-channels",
			Content: "Channels are typed conduits for communication between goroutines. You can send and receive values with the channel operator.",
			Metadata: map[string]interface{}{
				"category": "concurrency",
				"source":   "golang.org",
			},
		},
	}

	fmt.Printf("Adding %d documents to collection...\n", len(docs))
	ids, err := store.Add(ctx, collectionName, docs)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("✓ Added documents with IDs: %v\n", ids)

	// Check count
	count, err := store.Count(ctx, collectionName)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ Collection now has %d documents\n\n", count)

	// Example 4: Text Search (Semantic Search)
	fmt.Println("=== Example 4: Semantic Search by Text ===")

	query := "How does concurrent programming work in Go?"
	fmt.Printf("Query: %s\n\n", query)

	searchReq := agent.DefaultTextSearchRequest(collectionName, query)
	searchReq.TopK = 3
	searchReq.IncludeMetadata = true
	searchReq.IncludeContent = true

	results, err := store.SearchByText(ctx, searchReq)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Top %d results:\n", len(results))
	for i, result := range results {
		fmt.Printf("\n%d. [Score: %.4f] %s\n", i+1, result.Score, result.Document.ID)
		fmt.Printf("   Content: %s\n", result.Document.Content)
		if category, ok := result.Document.Metadata["category"]; ok {
			fmt.Printf("   Category: %s\n", category)
		}
	}
	fmt.Println()

	// Example 5: Vector Search with Pre-computed Embedding
	fmt.Println("=== Example 5: Search with Query Vector ===")

	// Generate embedding for a specific query
	queryText := "What are Go packages?"
	queryEmb, err := embedder.Embed(ctx, queryText)
	if err != nil {
		log.Fatal(err)
	}

	vectorSearchReq := agent.DefaultSearchRequest(collectionName, queryEmb)
	vectorSearchReq.TopK = 2

	vectorResults, err := store.Search(ctx, vectorSearchReq)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Query: %s\n", queryText)
	fmt.Printf("Results:\n")
	for i, result := range vectorResults {
		fmt.Printf("%d. [Score: %.4f] %s\n", i+1, result.Score, result.Document.Content[:60]+"...")
	}
	fmt.Println()

	// Example 6: Filtered Search
	fmt.Println("=== Example 6: Search with Metadata Filtering ===")

	filteredReq := agent.DefaultTextSearchRequest(collectionName, "Tell me about Go")
	filteredReq.TopK = 10
	filteredReq.Filter = map[string]interface{}{
		"category": "concurrency",
	}

	filteredResults, err := store.SearchByText(ctx, filteredReq)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Query: 'Tell me about Go' (filtered by category='concurrency')\n")
	fmt.Printf("Found %d results:\n", len(filteredResults))
	for i, result := range filteredResults {
		fmt.Printf("%d. %s\n", i+1, result.Document.Content[:80]+"...")
	}
	fmt.Println()

	// Example 7: Get Specific Documents
	fmt.Println("=== Example 7: Retrieve Specific Documents ===")

	retrieved, err := store.Get(ctx, collectionName, []string{"go-intro", "go-channels"})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Retrieved %d documents:\n", len(retrieved))
	for _, doc := range retrieved {
		fmt.Printf("- %s: %s\n", doc.ID, doc.Content[:50]+"...")
	}
	fmt.Println()

	// Example 8: Update Document
	fmt.Println("=== Example 8: Update Document ===")

	updateDoc := []*agent.VectorDocument{
		{
			ID:      "go-intro",
			Content: "Go (Golang) is a statically typed, compiled programming language designed at Google by Robert Griesemer, Rob Pike, and Ken Thompson.",
			Metadata: map[string]interface{}{
				"category": "introduction",
				"source":   "golang.org",
				"updated":  true,
			},
		},
	}

	err = store.Update(ctx, collectionName, updateDoc)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("✓ Document 'go-intro' updated")

	// Verify update
	updated, _ := store.Get(ctx, collectionName, []string{"go-intro"})
	if len(updated) > 0 {
		fmt.Printf("Updated content: %s\n", updated[0].Content[:80]+"...")
		if updated, ok := updated[0].Metadata["updated"]; ok {
			fmt.Printf("Updated flag: %v\n", updated)
		}
	}
	fmt.Println()

	// Example 9: Delete Documents
	fmt.Println("=== Example 9: Delete Specific Documents ===")

	beforeCount, _ := store.Count(ctx, collectionName)
	fmt.Printf("Documents before delete: %d\n", beforeCount)

	err = store.Delete(ctx, collectionName, []string{"go-packages"})
	if err != nil {
		log.Fatal(err)
	}

	afterCount, _ := store.Count(ctx, collectionName)
	fmt.Printf("Documents after delete: %d\n", afterCount)
	fmt.Println("✓ Document 'go-packages' deleted\n")

	// Example 10: Semantic Q&A System
	fmt.Println("=== Example 10: Semantic Q&A System ===")

	questions := []string{
		"What is a goroutine?",
		"How do I organize Go code?",
		"What makes Go special?",
	}

	for _, question := range questions {
		fmt.Printf("\nQ: %s\n", question)

		req := agent.DefaultTextSearchRequest(collectionName, question)
		req.TopK = 1

		results, err := store.SearchByText(ctx, req)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		if len(results) > 0 {
			fmt.Printf("A: %s\n", results[0].Document.Content)
			fmt.Printf("   (Confidence: %.2f%%)\n", results[0].Score*100)
		} else {
			fmt.Println("A: No relevant answer found")
		}
	}
	fmt.Println()

	// Example 11: Batch Embedding and Search
	fmt.Println("=== Example 11: Batch Operations ===")

	// Add multiple documents at once
	batchDocs := []*agent.VectorDocument{
		{
			Content: "Go has built-in testing support with the testing package.",
			Metadata: map[string]interface{}{
				"category": "testing",
			},
		},
		{
			Content: "Error handling in Go uses explicit error returns rather than exceptions.",
			Metadata: map[string]interface{}{
				"category": "errors",
			},
		},
		{
			Content: "Go modules provide dependency management for Go projects.",
			Metadata: map[string]interface{}{
				"category": "tools",
			},
		},
	}

	batchIDs, err := store.Add(ctx, collectionName, batchDocs)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("✓ Added %d documents in batch: %v\n", len(batchIDs), batchIDs)

	finalCount, _ := store.Count(ctx, collectionName)
	fmt.Printf("✓ Final collection size: %d documents\n\n", finalCount)

	// Example 12: Cleanup (Optional)
	fmt.Println("=== Example 12: Cleanup Operations ===")

	// Clear all documents but keep collection
	fmt.Println("To clear all documents: store.Clear(ctx, collectionName)")

	// Delete entire collection
	fmt.Println("To delete collection: store.DeleteCollection(ctx, collectionName)")

	// Uncomment to actually perform cleanup:
	/*
		err = store.Clear(ctx, collectionName)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("✓ Collection '%s' cleared\n", collectionName)

		err = store.DeleteCollection(ctx, collectionName)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("✓ Collection '%s' deleted\n", collectionName)
	*/

	fmt.Println("\n=== ChromaDB Examples Complete ===")
	fmt.Println("\nNotes:")
	fmt.Println("- Requires ChromaDB running: docker run -p 8000:8000 chromadb/chroma")
	fmt.Println("- Requires Ollama running: ollama serve")
	fmt.Println("- Install nomic-embed-text: ollama pull nomic-embed-text")
	fmt.Println("\nFor OpenAI embeddings instead of Ollama:")
	fmt.Println("  embedder, _ := agent.NewOpenAIEmbedding(agent.EmbeddingModelSmall, apiKey)")
}
