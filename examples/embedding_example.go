package main

import (
	"context"
	"fmt"
	"log"

	"github.com/taipm/go-deep-agent/agent"
)

func main() {
	ctx := context.Background()

	// Example 1: OpenAI Embedding Provider
	fmt.Println("=== Example 1: OpenAI Embeddings ===")
	openaiProvider, err := agent.NewOpenAIEmbedding(
		agent.EmbeddingModelSmall, // text-embedding-3-small (1536 dims)
		"YOUR_OPENAI_API_KEY",     // Replace with your API key
	)
	if err != nil {
		log.Fatal(err)
	}

	// Generate single embedding
	text := "The quick brown fox jumps over the lazy dog"
	embedding, err := openaiProvider.Embed(ctx, text)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Text: %s\n", text)
	fmt.Printf("Model: %s\n", openaiProvider.Model())
	fmt.Printf("Dimensions: %d\n", openaiProvider.Dimensions())
	fmt.Printf("Embedding (first 10 values): %v\n\n", embedding[:10])

	// Example 2: Batch Embeddings
	fmt.Println("=== Example 2: Batch Embeddings ===")
	texts := []string{
		"Machine learning is a subset of artificial intelligence",
		"Deep learning uses neural networks with multiple layers",
		"Natural language processing enables computers to understand text",
	}

	embeddings, err := openaiProvider.EmbedBatch(ctx, texts)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Generated %d embeddings\n", len(embeddings))
	for i, text := range texts {
		fmt.Printf("%d. %s\n", i+1, text)
		fmt.Printf("   Embedding dimensions: %d\n", len(embeddings[i]))
	}
	fmt.Println()

	// Example 3: Cosine Similarity
	fmt.Println("=== Example 3: Cosine Similarity ===")
	doc1 := "I love programming in Go"
	doc2 := "Go is a great programming language"
	doc3 := "The weather is nice today"

	emb1, _ := openaiProvider.Embed(ctx, doc1)
	emb2, _ := openaiProvider.Embed(ctx, doc2)
	emb3, _ := openaiProvider.Embed(ctx, doc3)

	sim12, _ := agent.CosineSimilarity(emb1, emb2)
	sim13, _ := agent.CosineSimilarity(emb1, emb3)
	sim23, _ := agent.CosineSimilarity(emb2, emb3)

	fmt.Printf("Document 1: %s\n", doc1)
	fmt.Printf("Document 2: %s\n", doc2)
	fmt.Printf("Document 3: %s\n\n", doc3)

	fmt.Printf("Similarity (Doc1 <-> Doc2): %.4f (similar topics)\n", sim12)
	fmt.Printf("Similarity (Doc1 <-> Doc3): %.4f (different topics)\n", sim13)
	fmt.Printf("Similarity (Doc2 <-> Doc3): %.4f (different topics)\n\n", sim23)

	// Example 4: Ollama Embedding Provider (Local)
	fmt.Println("=== Example 4: Ollama Embeddings (Local) ===")
	ollamaProvider, err := agent.NewOllamaEmbedding(
		agent.OllamaEmbeddingModelNomic, // nomic-embed-text (768 dims)
		"http://localhost:11434",        // Default Ollama server
	)
	if err != nil {
		log.Fatal(err)
	}

	// NOTE: Requires Ollama to be running with nomic-embed-text model installed
	// Run: ollama pull nomic-embed-text
	fmt.Printf("Model: %s\n", ollamaProvider.Model())
	fmt.Printf("Dimensions: %d\n", ollamaProvider.Dimensions())

	ollamaEmb, err := ollamaProvider.Embed(ctx, "Hello from Ollama!")
	if err != nil {
		fmt.Printf("Note: Ollama not available - %v\n", err)
	} else {
		fmt.Printf("Embedding (first 10 values): %v\n", ollamaEmb[:10])
	}
	fmt.Println()

	// Example 5: Vector Operations
	fmt.Println("=== Example 5: Vector Operations ===")
	vec1 := []float32{1, 2, 3}
	vec2 := []float32{4, 5, 6}

	dotProd, _ := agent.DotProduct(vec1, vec2)
	euclidean, _ := agent.EuclideanDistance(vec1, vec2)
	cosine, _ := agent.CosineSimilarity(vec1, vec2)

	fmt.Printf("Vector 1: %v\n", vec1)
	fmt.Printf("Vector 2: %v\n", vec2)
	fmt.Printf("Dot Product: %.4f\n", dotProd)
	fmt.Printf("Euclidean Distance: %.4f\n", euclidean)
	fmt.Printf("Cosine Similarity: %.4f\n\n", cosine)

	// Example 6: Vector Normalization
	fmt.Println("=== Example 6: Vector Normalization ===")
	vec := []float32{3, 4} // Magnitude is 5
	normalized := agent.NormalizeVector(vec)

	fmt.Printf("Original vector: %v (magnitude: 5)\n", vec)
	fmt.Printf("Normalized vector: %v\n", normalized)
	fmt.Printf("Expected: [0.6, 0.8]\n\n")

	// Example 7: Custom Configuration
	fmt.Println("=== Example 7: Custom Configuration ===")
	config := &agent.EmbeddingConfig{
		BatchSize:     50,   // Smaller batches
		Normalize:     true, // Normalize embeddings
		StripNewlines: true, // Clean text
	}

	customProvider, _ := agent.NewOpenAIEmbedding(
		agent.EmbeddingModelLarge, // text-embedding-3-large (3072 dims)
		"YOUR_OPENAI_API_KEY",
	)
	customProvider.WithConfig(config)

	multilineText := `This is a
	multiline text
	with newlines`

	customEmb, err := customProvider.Embed(ctx, multilineText)
	if err != nil {
		fmt.Printf("Note: Set your API key to run this example\n")
	} else {
		fmt.Printf("Text (with newlines): %q\n", multilineText)
		fmt.Printf("Embedding dimensions: %d\n", len(customEmb))
		fmt.Printf("Config: Normalize=%v, StripNewlines=%v\n\n",
			config.Normalize, config.StripNewlines)
	}

	// Example 8: Semantic Search Simulation
	fmt.Println("=== Example 8: Semantic Search Simulation ===")
	documents := []string{
		"Go is a statically typed, compiled programming language",
		"Python is a high-level, interpreted programming language",
		"The capital of France is Paris",
		"Machine learning is a field of artificial intelligence",
		"Go was designed at Google by Robert Griesemer, Rob Pike, and Ken Thompson",
	}

	query := "What programming language was created by Google?"

	// Embed all documents and query
	fmt.Printf("Query: %s\n\n", query)
	queryEmb, err := openaiProvider.Embed(ctx, query)
	if err != nil {
		fmt.Printf("Note: Set your API key to run semantic search\n")
		return
	}

	docEmbeddings, err := openaiProvider.EmbedBatch(ctx, documents)
	if err != nil {
		fmt.Printf("Note: Set your API key to run semantic search\n")
		return
	}

	// Calculate similarities
	type result struct {
		doc        string
		similarity float32
	}
	var results []result

	for i, docEmb := range docEmbeddings {
		sim, _ := agent.CosineSimilarity(queryEmb, docEmb)
		results = append(results, result{doc: documents[i], similarity: sim})
	}

	// Sort by similarity (simple bubble sort for demo)
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].similarity > results[i].similarity {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	// Print top 3 results
	fmt.Println("Top 3 most relevant documents:")
	for i := 0; i < 3 && i < len(results); i++ {
		fmt.Printf("%d. (%.4f) %s\n", i+1, results[i].similarity, results[i].doc)
	}
}
