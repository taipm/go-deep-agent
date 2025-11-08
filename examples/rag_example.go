package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/taipm/go-deep-agent/agent"
)

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	fmt.Println("=== RAG (Retrieval-Augmented Generation) Examples ===\n")

	// Example 1: Basic RAG
	basicRAG(apiKey)

	// Example 2: RAG with custom configuration
	ragWithConfig(apiKey)

	// Example 3: RAG with document metadata
	ragWithMetadata(apiKey)

	// Example 4: Custom retriever function
	customRetriever(apiKey)

	// Example 5: Getting retrieved documents
	inspectRetrievedDocs(apiKey)

	// Example 6: RAG for documentation Q&A
	documentationQA(apiKey)
}

// basicRAG demonstrates simple RAG usage
func basicRAG(apiKey string) {
	fmt.Println("1. Basic RAG Example")
	fmt.Println("-------------------")

	// Knowledge base about programming languages
	docs := []string{
		"Go is a statically typed, compiled programming language designed at Google by Robert Griesemer, Rob Pike, and Ken Thompson. Go is syntactically similar to C, but with memory safety, garbage collection, and structural typing.",
		"Python is an interpreted, high-level, general-purpose programming language created by Guido van Rossum and first released in 1991. Python's design philosophy emphasizes code readability with its notable use of significant indentation.",
		"Rust is a multi-paradigm, general-purpose programming language designed for performance and safety, especially safe concurrency. Rust is syntactically similar to C++, but can guarantee memory safety by using a borrow checker.",
		"JavaScript, often abbreviated JS, is a programming language that is one of the core technologies of the World Wide Web. It enables interactive web pages and is an essential part of web applications.",
	}

	// Create agent with RAG enabled
	ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithRAG(docs...).
		WithRAGTopK(2). // Retrieve top 2 most relevant chunks
		WithTemperature(0.7)

	ctx := context.Background()

	// Ask a question - relevant docs will be automatically retrieved
	response, err := ai.Ask(ctx, "Who created the Go programming language?")
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Question: Who created the Go programming language?\n")
	fmt.Printf("Answer: %s\n\n", response)
}

// ragWithConfig demonstrates custom RAG configuration
func ragWithConfig(apiKey string) {
	fmt.Println("2. RAG with Custom Configuration")
	fmt.Println("--------------------------------")

	longDoc := `
	The History of Programming Languages
	
	Programming languages have evolved significantly since the 1950s. Early languages like FORTRAN and COBOL 
	were designed for specific domains. In the 1970s, C was developed at Bell Labs by Dennis Ritchie, 
	becoming one of the most influential languages of all time.
	
	The 1990s saw the rise of object-oriented languages like Java and C++. Python, created by Guido van Rossum 
	in 1991, emphasized readability and simplicity. JavaScript, developed by Brendan Eich in 1995, became 
	the language of the web.
	
	In the 21st century, new languages emerged to address modern challenges. Go, announced by Google in 2009, 
	focused on simplicity and concurrency. Rust, first released in 2010, prioritized memory safety without 
	garbage collection. These modern languages continue to shape software development today.
	`

	// Custom configuration
	config := &agent.RAGConfig{
		ChunkSize:     300,           // Smaller chunks
		ChunkOverlap:  50,            // Small overlap
		TopK:          1,             // Only top result
		MinScore:      0.3,           // Minimum relevance score
		Separator:     "\n\n---\n\n", // Chunk separator
		IncludeScores: true,          // Show relevance scores
	}

	ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithRAG(longDoc).
		WithRAGConfig(config).
		WithTemperature(0.3)

	ctx := context.Background()

	response, err := ai.Ask(ctx, "When was Rust first released?")
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Question: When was Rust first released?\n")
	fmt.Printf("Answer: %s\n\n", response)
}

// ragWithMetadata demonstrates using documents with metadata
func ragWithMetadata(apiKey string) {
	fmt.Println("3. RAG with Document Metadata")
	fmt.Println("-----------------------------")

	// Documents with source metadata
	docs := []agent.Document{
		{
			Content: "The Go programming language was announced in November 2009 and became open source in 2012. It was designed at Google by Robert Griesemer, Rob Pike, and Ken Thompson.",
			Metadata: map[string]string{
				"source": "golang.org",
				"topic":  "go-history",
			},
		},
		{
			Content: "Go 1.0 was released in March 2012. The language has maintained backward compatibility since then, with the Go 1 compatibility guarantee.",
			Metadata: map[string]string{
				"source": "golang.org/doc/go1compat",
				"topic":  "go-versions",
			},
		},
		{
			Content: "Go is particularly well-suited for building microservices, CLI tools, and distributed systems. Major companies like Google, Uber, and Dropbox use Go in production.",
			Metadata: map[string]string{
				"source": "go-use-cases.md",
				"topic":  "go-adoption",
			},
		},
	}

	ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithRAGDocuments(docs...).
		WithRAGTopK(2).
		WithTemperature(0.5)

	ctx := context.Background()

	response, err := ai.Ask(ctx, "What companies use Go in production?")
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Question: What companies use Go in production?\n")
	fmt.Printf("Answer: %s\n", response)

	// Inspect retrieved documents
	retrieved := ai.GetLastRetrievedDocs()
	fmt.Printf("\nRetrieved %d documents:\n", len(retrieved))
	for i, doc := range retrieved {
		fmt.Printf("  %d. Source: %s (Score: %.2f)\n", i+1, doc.Metadata["source"], doc.Score)
	}
	fmt.Println()
}

// customRetriever demonstrates using a custom retriever function
func customRetriever(apiKey string) {
	fmt.Println("4. Custom Retriever Function")
	fmt.Println("----------------------------")

	// Simulate a database or external knowledge base
	knowledgeBase := map[string]string{
		"weather": "Today's weather in San Francisco is sunny with a high of 72°F.",
		"time":    "The current time is 3:45 PM PST.",
		"joke":    "Why do programmers prefer dark mode? Because light attracts bugs!",
	}

	// Custom retriever that looks up in our knowledge base
	retriever := func(query string) ([]agent.Document, error) {
		var docs []agent.Document

		// Simple keyword matching
		for key, content := range knowledgeBase {
			// In a real system, use proper similarity scoring
			if contains(query, key) {
				docs = append(docs, agent.Document{
					Content: content,
					Metadata: map[string]string{
						"source": "knowledge_base",
						"key":    key,
					},
					Score: 1.0,
				})
			}
		}

		return docs, nil
	}

	ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithRAGRetriever(retriever).
		WithTemperature(0.8)

	ctx := context.Background()

	response, err := ai.Ask(ctx, "Tell me a programming joke")
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Question: Tell me a programming joke\n")
	fmt.Printf("Answer: %s\n\n", response)
}

// inspectRetrievedDocs shows how to examine retrieved documents
func inspectRetrievedDocs(apiKey string) {
	fmt.Println("5. Inspecting Retrieved Documents")
	fmt.Println("---------------------------------")

	docs := []string{
		"Machine learning is a subset of artificial intelligence that focuses on building systems that learn from data.",
		"Deep learning is a specialized branch of machine learning that uses neural networks with multiple layers.",
		"Natural language processing (NLP) is a field of AI that deals with the interaction between computers and human language.",
		"Computer vision enables computers to derive meaningful information from digital images and videos.",
	}

	ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithRAG(docs...).
		WithRAGTopK(2).
		WithTemperature(0.6)

	ctx := context.Background()

	// Ask a question
	response, err := ai.Ask(ctx, "What is deep learning?")
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Question: What is deep learning?\n")
	fmt.Printf("Answer: %s\n\n", response)

	// Inspect what was retrieved
	retrieved := ai.GetLastRetrievedDocs()
	fmt.Printf("Retrieved %d documents:\n", len(retrieved))
	for i, doc := range retrieved {
		fmt.Printf("\nDocument %d (Score: %.3f):\n", i+1, doc.Score)
		fmt.Printf("%s\n", doc.Content)
	}
	fmt.Println()
}

// documentationQA demonstrates RAG for documentation Q&A
func documentationQA(apiKey string) {
	fmt.Println("6. Documentation Q&A System")
	fmt.Println("---------------------------")

	// Simulate API documentation
	apiDocs := []agent.Document{
		{
			Content: "POST /api/users - Creates a new user. Required fields: name (string), email (string). Returns: user object with id.",
			Metadata: map[string]string{
				"endpoint": "/api/users",
				"method":   "POST",
			},
		},
		{
			Content: "GET /api/users/:id - Retrieves a user by ID. Returns: user object or 404 if not found.",
			Metadata: map[string]string{
				"endpoint": "/api/users/:id",
				"method":   "GET",
			},
		},
		{
			Content: "DELETE /api/users/:id - Deletes a user by ID. Requires authentication token. Returns: 204 on success.",
			Metadata: map[string]string{
				"endpoint": "/api/users/:id",
				"method":   "DELETE",
			},
		},
		{
			Content: "Authentication: All API requests require a Bearer token in the Authorization header. Tokens expire after 24 hours.",
			Metadata: map[string]string{
				"topic": "authentication",
			},
		},
	}

	ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithRAGDocuments(apiDocs...).
		WithRAGTopK(2).
		WithTemperature(0.3). // Low temperature for factual answers
		WithSystem("You are a helpful API documentation assistant. Answer questions based on the provided documentation. Be concise and accurate.")

	ctx := context.Background()

	questions := []string{
		"How do I create a new user?",
		"What authentication is required?",
		"How do I delete a user?",
	}

	for _, question := range questions {
		response, err := ai.Ask(ctx, question)
		if err != nil {
			log.Printf("Error: %v\n", err)
			continue
		}

		fmt.Printf("Q: %s\n", question)
		fmt.Printf("A: %s\n\n", response)

		// Small delay to avoid rate limits
		time.Sleep(500 * time.Millisecond)
	}
}

// Helper function for simple string containment
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
