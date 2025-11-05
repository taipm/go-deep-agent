package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/taipm/go-deep-agent/agent"
)

// Example 1: Simple JSON Mode
// JSON Mode forces the model to return valid JSON,
// but you need to instruct it what structure to use
func example1_JSONMode() {
	fmt.Println("=== Example 1: JSON Mode ===")

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is not set")
	}

	ctx := context.Background()

	// Create builder with JSON mode enabled
	response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithJSONMode().
		WithSystem("You are a helpful assistant. Always respond with valid JSON containing 'answer' and 'confidence' fields.").
		Ask(ctx, "What is the capital of France?")

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Response:", response)

	// Parse the JSON response
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		log.Printf("Failed to parse JSON: %v", err)
	} else {
		fmt.Printf("Parsed: answer=%v, confidence=%v\n", result["answer"], result["confidence"])
	}
	fmt.Println()
}

// Example 2: JSON Schema - Weather Response
// JSON Schema ensures the model follows an exact structure
func example2_WeatherSchema() {
	fmt.Println("=== Example 2: Weather JSON Schema ===")

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is not set")
	}

	ctx := context.Background()

	// Define a JSON schema for weather response
	// Note: In strict mode, ALL properties must be in required array
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"location": map[string]interface{}{
				"type":        "string",
				"description": "The location name",
			},
			"temperature": map[string]interface{}{
				"type":        "number",
				"description": "Temperature in Celsius",
			},
			"condition": map[string]interface{}{
				"type":        "string",
				"description": "Weather condition (e.g., sunny, cloudy, rainy)",
			},
			"humidity": map[string]interface{}{
				"type":        "number",
				"description": "Humidity percentage",
			},
		},
		"required":             []string{"location", "temperature", "condition", "humidity"},
		"additionalProperties": false,
	}

	// Create builder with JSON schema
	response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithJSONSchema("weather_response", "Weather information for a location", schema, true).
		Ask(ctx, "What's the weather like in Tokyo right now?")

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Response:", response)

	// Parse the structured response
	var weather map[string]interface{}
	if err := json.Unmarshal([]byte(response), &weather); err != nil {
		log.Printf("Failed to parse JSON: %v", err)
	} else {
		fmt.Printf("Location: %v\n", weather["location"])
		fmt.Printf("Temperature: %.1f°C\n", weather["temperature"])
		fmt.Printf("Condition: %v\n", weather["condition"])
		if humidity, ok := weather["humidity"]; ok {
			fmt.Printf("Humidity: %.0f%%\n", humidity)
		}
	}
	fmt.Println()
}

// Example 3: JSON Schema - Structured Data Extraction
// Extract structured information from unstructured text
func example3_DataExtraction() {
	fmt.Println("=== Example 3: Data Extraction with JSON Schema ===")

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is not set")
	}

	ctx := context.Background()

	// Define schema for person information
	// Note: In strict mode, ALL properties must be in required array
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Full name of the person",
			},
			"age": map[string]interface{}{
				"type":        "integer",
				"description": "Age in years",
			},
			"occupation": map[string]interface{}{
				"type":        "string",
				"description": "Current occupation",
			},
			"location": map[string]interface{}{
				"type":        "string",
				"description": "City and country of residence",
			},
			"skills": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
				},
				"description": "List of skills",
			},
		},
		"required":             []string{"name", "age", "occupation", "location", "skills"},
		"additionalProperties": false,
	}

	text := `
	John Smith is a 32-year-old software engineer living in San Francisco, USA.
	He specializes in Go, Python, and cloud infrastructure. In his free time,
	he enjoys hiking and photography.
	`

	response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithJSONSchema("person_info", "Structured person information", schema, true).
		Ask(ctx, fmt.Sprintf("Extract structured information from this text: %s", text))

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Response:", response)

	// Parse the structured data
	var person map[string]interface{}
	if err := json.Unmarshal([]byte(response), &person); err != nil {
		log.Printf("Failed to parse JSON: %v", err)
	} else {
		fmt.Printf("\n📋 Extracted Information:\n")
		fmt.Printf("  Name: %v\n", person["name"])
		fmt.Printf("  Age: %.0f\n", person["age"])
		fmt.Printf("  Occupation: %v\n", person["occupation"])
		if location, ok := person["location"]; ok {
			fmt.Printf("  Location: %v\n", location)
		}
		if skills, ok := person["skills"].([]interface{}); ok {
			fmt.Printf("  Skills: %v\n", skills)
		}
	}
	fmt.Println()
}

// Example 4: JSON Schema - Complex Nested Structure
// Demonstrate complex nested objects and arrays
func example4_NestedStructure() {
	fmt.Println("=== Example 4: Complex Nested JSON Schema ===")

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is not set")
	}

	ctx := context.Background()

	// Define schema for a book review with nested structures
	// Note: In strict mode, nested objects also need additionalProperties: false
	// and all properties must be required
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"book": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"title": map[string]interface{}{
						"type": "string",
					},
					"author": map[string]interface{}{
						"type": "string",
					},
					"year": map[string]interface{}{
						"type": "integer",
					},
				},
				"required":             []string{"title", "author", "year"},
				"additionalProperties": false,
			},
			"review": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"rating": map[string]interface{}{
						"type":    "integer",
						"minimum": 1,
						"maximum": 5,
					},
					"summary": map[string]interface{}{
						"type": "string",
					},
					"pros": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "string",
						},
					},
					"cons": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "string",
						},
					},
				},
				"required":             []string{"rating", "summary", "pros", "cons"},
				"additionalProperties": false,
			},
			"recommend": map[string]interface{}{
				"type": "boolean",
			},
		},
		"required":             []string{"book", "review", "recommend"},
		"additionalProperties": false,
	}

	response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithJSONSchema("book_review", "Structured book review", schema, true).
		Ask(ctx, "Write a review for '1984' by George Orwell")

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Response:", response)

	// Parse and pretty print the structured review
	var review map[string]interface{}
	if err := json.Unmarshal([]byte(response), &review); err != nil {
		log.Printf("Failed to parse JSON: %v", err)
	} else {
		prettyJSON, _ := json.MarshalIndent(review, "", "  ")
		fmt.Printf("\n📚 Structured Review:\n%s\n", prettyJSON)
	}
	fmt.Println()
}

func main() {
	fmt.Println("=== OpenAI JSON Schema Examples ===\n")

	// Run all examples
	example1_JSONMode()
	example2_WeatherSchema()
	example3_DataExtraction()
	example4_NestedStructure()

	fmt.Println("✅ All examples completed!")
}
