package main

import (
	"context"
	"fmt"
	"log"
	"os"

	agent "github.com/taipm/go-deep-agent/agent"
)

func main() {
	fmt.Println("=== Vector RAG (Retrieval-Augmented Generation) Examples ===\n")

	// Prerequisites:
	// 1. Set OPENAI_API_KEY environment variable
	// 2. Start ChromaDB: docker run -p 8000:8000 chromadb/chroma
	//    OR Start Qdrant: docker run -p 6333:6333 qdrant/qdrant
	// 3. Start Ollama: ollama serve
	// 4. Pull model: ollama pull nomic-embed-text

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable not set")
	}

	ctx := context.Background()

	// Example 1: Setup Vector RAG with ChromaDB
	fmt.Println("Example 1: Setup Vector RAG System")
	fmt.Println("-----------------------------------")

	// Create embedding provider (using Ollama for free embeddings)
	embedding, err := agent.NewOllamaEmbedding(
		"http://localhost:11434",
		"nomic-embed-text",
	)
	if err != nil {
		log.Fatalf("Failed to create embedding provider: %v", err)
	}

	// Create vector store (using ChromaDB)
	chromaStore, err := agent.NewChromaStore("http://localhost:8000")
	if err != nil {
		log.Fatalf("Failed to create ChromaDB store: %v", err)
	}
	chromaStore.WithEmbedding(embedding)

	// Create collection
	collectionConfig := &agent.CollectionConfig{
		Name:           "knowledge-base",
		Description:    "Company knowledge base",
		Dimension:      768, // nomic-embed-text dimension
		DistanceMetric: agent.DistanceMetricCosine,
	}

	err = chromaStore.CreateCollection(ctx, "knowledge-base", collectionConfig)
	if err != nil {
		log.Printf("Warning: Collection might already exist: %v\n", err)
	}

	fmt.Println("✓ Vector store configured with ChromaDB")
	fmt.Println("✓ Using Ollama embeddings (nomic-embed-text)")
	fmt.Println()

	// Example 2: Add knowledge base documents
	fmt.Println("Example 2: Add Documents to Vector Store")
	fmt.Println("-----------------------------------------")

	knowledgeDocs := []string{
		"Our company was founded in 2020 and specializes in AI-powered solutions for enterprise customers.",
		"We offer 24/7 customer support via email at support@example.com and phone at 1-800-SUPPORT.",
		"Our refund policy allows full refunds within 30 days of purchase for any reason.",
		"Premium plans include advanced analytics, priority support, and custom integrations.",
		"Security is our top priority. We use AES-256 encryption and are SOC 2 Type II certified.",
		"Our API supports REST and GraphQL with rate limits of 1000 requests per hour for free tier.",
		"Enterprise plans start at $999/month and include dedicated account management.",
		"We support integrations with Slack, Microsoft Teams, Salesforce, and 50+ other platforms.",
		"Our data centers are located in US-East, US-West, EU-West, and Asia-Pacific regions.",
		"All customer data is backed up hourly and retained for 90 days with point-in-time recovery.",
	}

	// Create agent with vector RAG
	aiAgent := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithVectorRAG(embedding, chromaStore, "knowledge-base")

	// Add documents to vector store
	ids, err := aiAgent.AddDocumentsToVector(ctx, knowledgeDocs...)
	if err != nil {
		log.Fatalf("Failed to add documents: %v", err)
	}

	fmt.Printf("✓ Added %d documents to vector store\n", len(ids))
	fmt.Println()

	// Example 3: Simple RAG query
	fmt.Println("Example 3: Simple RAG Query")
	fmt.Println("----------------------------")

	response, err := aiAgent.Ask(ctx, "What is your refund policy?")
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}

	fmt.Println("Q: What is your refund policy?")
	fmt.Printf("A: %s\n", response)

	// Show retrieved documents
	retrievedDocs := aiAgent.GetLastRetrievedDocs()
	fmt.Printf("\n📚 Retrieved %d relevant documents:\n", len(retrievedDocs))
	for i, doc := range retrievedDocs {
		fmt.Printf("  %d. [Score: %.3f] %s\n", i+1, doc.Score, doc.Content)
	}
	fmt.Println()

	// Example 4: Multi-turn conversation with RAG
	fmt.Println("Example 4: Multi-turn Conversation")
	fmt.Println("-----------------------------------")

	chatAgent := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithVectorRAG(embedding, chromaStore, "knowledge-base").
		WithMemory() // Enable conversation memory

	questions := []string{
		"How can I contact support?",
		"What payment plans do you offer?",
		"Tell me about your security measures.",
	}

	for i, question := range questions {
		fmt.Printf("\nTurn %d:\n", i+1)
		fmt.Printf("Q: %s\n", question)

		answer, err := chatAgent.Ask(ctx, question)
		if err != nil {
			log.Printf("Query failed: %v\n", err)
			continue
		}

		fmt.Printf("A: %s\n", answer)

		// Show top retrieved doc
		docs := chatAgent.GetLastRetrievedDocs()
		if len(docs) > 0 {
			fmt.Printf("   └─ Source: \"%s\" (score: %.3f)\n", docs[0].Content[:60]+"...", docs[0].Score)
		}
	}
	fmt.Println()

	// Example 5: Custom TopK and MinScore
	fmt.Println("Example 5: Custom Retrieval Parameters")
	fmt.Println("---------------------------------------")

	customAgent := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithVectorRAG(embedding, chromaStore, "knowledge-base").
		WithRAGTopK(5). // Retrieve top 5 documents
		WithRAGConfig(&agent.RAGConfig{
			TopK:          5,
			MinScore:      0.7, // Only use high-confidence results
			Separator:     "\n\n",
			IncludeScores: true,
		})

	response, err = customAgent.Ask(ctx, "Tell me about your company and services")
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}

	fmt.Println("Q: Tell me about your company and services")
	fmt.Printf("A: %s\n\n", response)

	docs := customAgent.GetLastRetrievedDocs()
	fmt.Printf("Retrieved %d documents (MinScore: 0.7):\n", len(docs))
	for i, doc := range docs {
		fmt.Printf("  %d. [%.3f] %s\n", i+1, doc.Score, doc.Content[:60]+"...")
	}
	fmt.Println()

	// Example 6: Using Qdrant instead of ChromaDB
	fmt.Println("Example 6: Switch to Qdrant")
	fmt.Println("---------------------------")

	// Create Qdrant store
	qdrantStore, err := agent.NewQdrantStore("http://localhost:6333")
	if err != nil {
		log.Printf("Qdrant not available: %v\n", err)
	} else {
		qdrantStore.WithEmbedding(embedding)

		// Create collection
		err = qdrantStore.CreateCollection(ctx, "knowledge-base", collectionConfig)
		if err != nil {
			log.Printf("Collection exists or error: %v\n", err)
		}

		// Create agent with Qdrant
		qdrantAgent := agent.NewOpenAI("gpt-4o-mini", apiKey).
			WithVectorRAG(embedding, qdrantStore, "knowledge-base")

		// Add same documents to Qdrant
		_, err = qdrantAgent.AddDocumentsToVector(ctx, knowledgeDocs...)
		if err != nil {
			log.Printf("Failed to add to Qdrant: %v\n", err)
		} else {
			fmt.Println("✓ Switched to Qdrant vector store")

			// Query with Qdrant
			response, err := qdrantAgent.Ask(ctx, "What regions are your data centers in?")
			if err != nil {
				log.Printf("Query failed: %v\n", err)
			} else {
				fmt.Println("Q: What regions are your data centers in?")
				fmt.Printf("A: %s\n", response)
			}
		}
	}
	fmt.Println()

	// Example 7: Adding documents with metadata
	fmt.Println("Example 7: Documents with Metadata")
	fmt.Println("-----------------------------------")

	metadataAgent := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithVectorRAG(embedding, chromaStore, "knowledge-base")

	// Add documents with rich metadata
	vectorDocs := []*agent.VectorDocument{
		{
			Content: "Python is our primary backend language, using FastAPI and Django frameworks.",
			Metadata: map[string]interface{}{
				"category":   "technical",
				"department": "engineering",
				"topic":      "programming",
			},
		},
		{
			Content: "We use PostgreSQL for relational data and Redis for caching.",
			Metadata: map[string]interface{}{
				"category":   "technical",
				"department": "engineering",
				"topic":      "database",
			},
		},
		{
			Content: "Our hiring process includes technical interview, system design, and cultural fit assessment.",
			Metadata: map[string]interface{}{
				"category":   "hr",
				"department": "human resources",
				"topic":      "recruitment",
			},
		},
	}

	ids, err = metadataAgent.AddVectorDocuments(ctx, vectorDocs...)
	if err != nil {
		log.Printf("Failed to add documents: %v\n", err)
	} else {
		fmt.Printf("✓ Added %d documents with metadata\n", len(ids))

		// Query technical information
		response, err := metadataAgent.Ask(ctx, "What technologies do you use for backend?")
		if err != nil {
			log.Printf("Query failed: %v\n", err)
		} else {
			fmt.Println("\nQ: What technologies do you use for backend?")
			fmt.Printf("A: %s\n", response)

			docs := metadataAgent.GetLastRetrievedDocs()
			if len(docs) > 0 {
				fmt.Printf("\n📋 Metadata from top result:\n")
				for k, v := range docs[0].Metadata {
					fmt.Printf("   %s: %s\n", k, v)
				}
			}
		}
	}
	fmt.Println()

	// Example 8: Comparison with traditional RAG
	fmt.Println("Example 8: Vector vs Traditional RAG")
	fmt.Println("-------------------------------------")

	// Traditional TF-IDF RAG
	traditionalAgent := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithRAG(knowledgeDocs...).
		WithRAGTopK(3)

	fmt.Println("Traditional RAG (TF-IDF):")
	response1, _ := traditionalAgent.Ask(ctx, "security encryption")
	fmt.Printf("  Query: 'security encryption'\n")
	fmt.Printf("  Top doc score: %.3f\n", traditionalAgent.GetLastRetrievedDocs()[0].Score)

	// Vector RAG
	vectorAgent := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithVectorRAG(embedding, chromaStore, "knowledge-base").
		WithRAGTopK(3)

	fmt.Println("\nVector RAG (Semantic Search):")
	response2, _ := vectorAgent.Ask(ctx, "security encryption")
	fmt.Printf("  Query: 'security encryption'\n")
	fmt.Printf("  Top doc score: %.3f\n", vectorAgent.GetLastRetrievedDocs()[0].Score)

	fmt.Println("\n💡 Vector RAG provides better semantic understanding")
	fmt.Println("   Traditional: Keyword matching")
	fmt.Println("   Vector: Meaning-based retrieval")
	fmt.Println()

	// Cleanup
	fmt.Println("=== Summary ===")
	fmt.Println("✓ Vector RAG enables semantic search with embeddings")
	fmt.Println("✓ Supports ChromaDB, Qdrant, and other vector stores")
	fmt.Println("✓ Better retrieval accuracy than keyword-based methods")
	fmt.Println("✓ Configurable TopK, MinScore, and metadata filtering")
	fmt.Println("✓ Seamless integration with existing RAG system")
	fmt.Println()

	// Optional cleanup (uncomment to delete)
	// chromaStore.DeleteCollection(ctx, "knowledge-base")
	// fmt.Println("✓ Collection deleted")

	_, _ = response1, response2 // Suppress unused warnings
}
