// Package main demonstrates multimodal capabilities with GPT-4 Vision.
//
// GPT-4 Vision can analyze images combined with text prompts.
// It supports image URLs, local files, and base64-encoded images.
//
// To run: OPENAI_API_KEY=your-key go run examples/builder_multimodal.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/taipm/go-deep-agent/agent"
)

func main() {
	// Get API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	ctx := context.Background()

	// Example 1: Image Description (URL)
	fmt.Println("=== Example 1: Describe an Image from URL ===")
	example1ImageDescription(ctx, apiKey)
	fmt.Println()

	// Example 2: Multiple Images Comparison
	fmt.Println("=== Example 2: Compare Multiple Images ===")
	example2MultipleImages(ctx, apiKey)
	fmt.Println()

	// Example 3: Detail Levels (Low vs High)
	fmt.Println("=== Example 3: Detail Levels ===")
	example3DetailLevels(ctx, apiKey)
	fmt.Println()

	// Example 4: OCR and Text Extraction
	fmt.Println("=== Example 4: OCR - Extract Text from Image ===")
	example4OCR(ctx, apiKey)
	fmt.Println()

	// Example 5: Image Analysis with Local File
	fmt.Println("=== Example 5: Analyze Local Image File ===")
	example5LocalFile(ctx, apiKey)
	fmt.Println()

	// Example 6: Chart/Graph Analysis
	fmt.Println("=== Example 6: Analyze Chart or Graph ===")
	example6ChartAnalysis(ctx, apiKey)
	fmt.Println()

	// Example 7: Multiple Images in Conversation
	fmt.Println("=== Example 7: Conversation with Images ===")
	example7Conversation(ctx, apiKey)
	fmt.Println()
}

// Example 1: Basic image description
func example1ImageDescription(ctx context.Context, apiKey string) {
	// Image: Sample landscape photo
	imageURL := "https://upload.wikimedia.org/wikipedia/commons/thumb/d/dd/Gfp-wisconsin-madison-the-nature-boardwalk.jpg/2560px-Gfp-wisconsin-madison-the-nature-boardwalk.jpg"

	response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithImage(imageURL).
		Ask(ctx, "What do you see in this image? Describe it in detail.")

	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	fmt.Println(response)
}

// Example 2: Compare multiple images
func example2MultipleImages(ctx context.Context, apiKey string) {
	// Two different landscape photos
	image1 := "https://upload.wikimedia.org/wikipedia/commons/thumb/d/dd/Gfp-wisconsin-madison-the-nature-boardwalk.jpg/2560px-Gfp-wisconsin-madison-the-nature-boardwalk.jpg"
	image2 := "https://upload.wikimedia.org/wikipedia/commons/thumb/3/3f/Alps_in_Switzerland.jpg/1920px-Alps_in_Switzerland.jpg"

	response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithImage(image1).
		WithImage(image2).
		Ask(ctx, "Compare these two images. What are the similarities and differences?")

	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	fmt.Println(response)
}

// Example 3: Using different detail levels
func example3DetailLevels(ctx context.Context, apiKey string) {
	imageURL := "https://upload.wikimedia.org/wikipedia/commons/thumb/d/dd/Gfp-wisconsin-madison-the-nature-boardwalk.jpg/2560px-Gfp-wisconsin-madison-the-nature-boardwalk.jpg"

	// Low detail (faster, cheaper)
	fmt.Println("--- Low Detail (faster, uses fewer tokens) ---")
	response1, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithImageURL(imageURL, agent.ImageDetailLow).
		Ask(ctx, "What's in this image?")

	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Println(response1)
	}

	fmt.Println()

	// High detail (more detailed analysis)
	fmt.Println("--- High Detail (more detailed analysis) ---")
	response2, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithImageURL(imageURL, agent.ImageDetailHigh).
		Ask(ctx, "Describe this image in great detail.")

	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Println(response2)
	}
}

// Example 4: OCR - Extract text from image
func example4OCR(ctx context.Context, apiKey string) {
	// Image with text (e.g., a sign, document, or screenshot)
	// Using a Wikipedia image with text
	imageURL := "https://upload.wikimedia.org/wikipedia/commons/thumb/7/78/World_Flags.png/1920px-World_Flags.png"

	response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithImageURL(imageURL, agent.ImageDetailHigh).
		Ask(ctx, "Extract all text you can see in this image. List each line separately.")

	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	fmt.Println(response)
}

// Example 5: Analyze local image file
func example5LocalFile(ctx context.Context, apiKey string) {
	// NOTE: This requires a local image file
	// For demonstration, we'll just show the API usage
	// Uncomment and modify the path to test with your own image

	/*
		response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
			WithImageFile("./path/to/your/image.jpg", agent.ImageDetailHigh).
			Ask(ctx, "What's in this image?")

		if err != nil {
			log.Printf("Error: %v\n", err)
			return
		}

		fmt.Println(response)
	*/

	fmt.Println("(Example skipped - requires local file)")
	fmt.Println("Usage:")
	fmt.Println("  builder.WithImageFile(\"./photo.jpg\", agent.ImageDetailHigh)")
}

// Example 6: Analyze charts and graphs
func example6ChartAnalysis(ctx context.Context, apiKey string) {
	// Image of a chart or graph
	chartURL := "https://upload.wikimedia.org/wikipedia/commons/thumb/5/5e/World_population_growth_rates_%281950-2050%29.svg/1920px-World_population_growth_rates_%281950-2050%29.svg.png"

	response, err := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithImageURL(chartURL, agent.ImageDetailHigh).
		Ask(ctx, "Analyze this chart. What are the key trends and insights?")

	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	fmt.Println(response)
}

// Example 7: Conversation with images
func example7Conversation(ctx context.Context, apiKey string) {
	builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithMemory()

	image1 := "https://upload.wikimedia.org/wikipedia/commons/thumb/d/dd/Gfp-wisconsin-madison-the-nature-boardwalk.jpg/2560px-Gfp-wisconsin-madison-the-nature-boardwalk.jpg"

	// First message with image
	response1, err := builder.
		WithImage(image1).
		Ask(ctx, "What season is shown in this image?")

	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Q1: What season is shown in this image?")
	fmt.Println("A1:", response1)
	fmt.Println()

	// Follow-up without image (uses conversation memory)
	response2, err := builder.Ask(ctx, "What activities could you do there?")

	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Q2: What activities could you do there?")
	fmt.Println("A2:", response2)
	fmt.Println()

	// Add another image
	image2 := "https://upload.wikimedia.org/wikipedia/commons/thumb/3/3f/Alps_in_Switzerland.jpg/1920px-Alps_in_Switzerland.jpg"

	response3, err := builder.
		WithImage(image2).
		Ask(ctx, "How is this location different from the previous one?")

	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Q3: How is this location different from the previous one?")
	fmt.Println("A3:", response3)
}
