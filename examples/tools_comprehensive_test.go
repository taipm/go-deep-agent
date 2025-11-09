package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/taipm/go-deep-agent/agent"
	"github.com/taipm/go-deep-agent/agent/tools"
)

// Comprehensive test of all 4 built-in tools
// This demonstrates real-world usage scenarios

func main() {
	fmt.Println("=== 🧪 COMPREHENSIVE BUILT-IN TOOLS TEST ===\n")

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("❌ OPENAI_API_KEY environment variable not set")
	}

	ctx := context.Background()

	// Test 1: Individual Tool Tests
	fmt.Println("📋 TEST 1: Individual Tool Functionality\n")
	testFileSystemTool(ctx, apiKey)
	testHTTPRequestTool(ctx, apiKey)
	testDateTimeTool(ctx, apiKey)
	testMathTool(ctx, apiKey)

	// Test 2: Multi-Tool Integration
	fmt.Println("\n📋 TEST 2: Multi-Tool Integration\n")
	testMultiToolIntegration(ctx, apiKey)

	// Test 3: Complex Real-World Scenario
	fmt.Println("\n📋 TEST 3: Complex Real-World Scenario\n")
	testRealWorldScenario(ctx, apiKey)

	// Test 4: Error Handling
	fmt.Println("\n📋 TEST 4: Error Handling & Edge Cases\n")
	testErrorHandling(ctx, apiKey)

	// Test 5: Performance Benchmarks
	fmt.Println("\n📋 TEST 5: Performance Benchmarks\n")
	testPerformance(ctx, apiKey)

	fmt.Println("\n✅ ALL TESTS COMPLETED!")
}

// Test 1.1: FileSystemTool
func testFileSystemTool(ctx context.Context, apiKey string) {
	fmt.Println("🗂️  Testing FileSystemTool...")

	fsTool := tools.NewFileSystemTool()
	ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithTool(fsTool).
		WithAutoExecute(true)

	testCases := []struct {
		name  string
		query string
	}{
		{"Write file", "Create a file called test_data.txt with content 'Hello from AI Agent'"},
		{"Read file", "Read the file test_data.txt"},
		{"Check exists", "Does the file test_data.txt exist?"},
		{"List directory", "List all files in the current directory"},
	}

	for _, tc := range testCases {
		fmt.Printf("  ▶ %s: ", tc.name)
		start := time.Now()
		response, err := ai.Ask(ctx, tc.query)
		duration := time.Since(start)

		if err != nil {
			fmt.Printf("❌ FAILED (%v) - %v\n", duration, err)
		} else {
			fmt.Printf("✅ PASSED (%v)\n", duration)
			if len(response) > 100 {
				fmt.Printf("    Response: %s...\n", response[:100])
			} else {
				fmt.Printf("    Response: %s\n", response)
			}
		}
	}

	// Cleanup
	os.Remove("test_data.txt")
	fmt.Println()
}

// Test 1.2: HTTPRequestTool
func testHTTPRequestTool(ctx context.Context, apiKey string) {
	fmt.Println("🌐 Testing HTTPRequestTool...")

	httpTool := tools.NewHTTPRequestTool()
	ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithTool(httpTool).
		WithAutoExecute(true)

	testCases := []struct {
		name  string
		query string
	}{
		{"GET request", "Fetch data from https://jsonplaceholder.typicode.com/todos/1"},
		{"GET user", "Get user data from https://jsonplaceholder.typicode.com/users/1"},
		{"GET posts", "Fetch the first post from https://jsonplaceholder.typicode.com/posts/1"},
	}

	for _, tc := range testCases {
		fmt.Printf("  ▶ %s: ", tc.name)
		start := time.Now()
		response, err := ai.Ask(ctx, tc.query)
		duration := time.Since(start)

		if err != nil {
			fmt.Printf("❌ FAILED (%v) - %v\n", duration, err)
		} else {
			fmt.Printf("✅ PASSED (%v)\n", duration)
			if len(response) > 100 {
				fmt.Printf("    Response: %s...\n", response[:100])
			}
		}
	}
	fmt.Println()
}

// Test 1.3: DateTimeTool
func testDateTimeTool(ctx context.Context, apiKey string) {
	fmt.Println("📅 Testing DateTimeTool...")

	dtTool := tools.NewDateTimeTool()
	ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithTool(dtTool).
		WithAutoExecute(true)

	testCases := []struct {
		name  string
		query string
	}{
		{"Current time", "What is the current time in UTC?"},
		{"Day of week", "What day of the week is Christmas 2025 (December 25)?"},
		{"Timezone convert", "What time is it in Tokyo right now?"},
		{"Date difference", "How many days until New Year 2026?"},
	}

	for _, tc := range testCases {
		fmt.Printf("  ▶ %s: ", tc.name)
		start := time.Now()
		response, err := ai.Ask(ctx, tc.query)
		duration := time.Since(start)

		if err != nil {
			fmt.Printf("❌ FAILED (%v) - %v\n", duration, err)
		} else {
			fmt.Printf("✅ PASSED (%v)\n", duration)
			fmt.Printf("    Response: %s\n", response)
		}
	}
	fmt.Println()
}

// Test 1.4: MathTool
func testMathTool(ctx context.Context, apiKey string) {
	fmt.Println("🧮 Testing MathTool...")

	mathTool := tools.NewMathTool()
	ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithTool(mathTool).
		WithAutoExecute(true)

	testCases := []struct {
		name  string
		query string
	}{
		{"Expression eval", "Calculate: 2 * (3 + 4) + sqrt(16)"},
		{"Statistics", "What is the average of these numbers: 10, 20, 30, 40, 50?"},
		{"Equation solve", "Solve this equation: x + 15 = 42"},
		{"Unit convert", "Convert 100 kilometers to meters"},
		{"Random gen", "Generate a random number between 1 and 100"},
		{"Complex expr", "Calculate: sin(3.14159/2) + pow(2, 3) - sqrt(9)"},
	}

	for _, tc := range testCases {
		fmt.Printf("  ▶ %s: ", tc.name)
		start := time.Now()
		response, err := ai.Ask(ctx, tc.query)
		duration := time.Since(start)

		if err != nil {
			fmt.Printf("❌ FAILED (%v) - %v\n", duration, err)
		} else {
			fmt.Printf("✅ PASSED (%v)\n", duration)
			fmt.Printf("    Response: %s\n", response)
		}
	}
	fmt.Println()
}

// Test 2: Multi-Tool Integration
func testMultiToolIntegration(ctx context.Context, apiKey string) {
	fmt.Println("🔗 Testing Multi-Tool Integration...")

	// Create all tools
	fsTool := tools.NewFileSystemTool()
	httpTool := tools.NewHTTPRequestTool()
	dtTool := tools.NewDateTimeTool()
	mathTool := tools.NewMathTool()

	ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithTools(fsTool, httpTool, dtTool, mathTool).
		WithAutoExecute(true).
		WithMaxToolRounds(10)

	testCases := []struct {
		name  string
		query string
	}{
		{
			"Fetch + Save",
			"Fetch data from https://jsonplaceholder.typicode.com/posts/1 and save it to post.json",
		},
		{
			"Math + Time",
			"Get current time and calculate how many seconds in 24 hours",
		},
		{
			"Complex workflow",
			"Calculate the mean of [5, 10, 15, 20, 25], then create a file results.txt with the answer",
		},
	}

	for _, tc := range testCases {
		fmt.Printf("  ▶ %s: ", tc.name)
		start := time.Now()
		response, err := ai.Ask(ctx, tc.query)
		duration := time.Since(start)

		if err != nil {
			fmt.Printf("❌ FAILED (%v) - %v\n", duration, err)
		} else {
			fmt.Printf("✅ PASSED (%v)\n", duration)
			fmt.Printf("    Response: %s\n", response)
		}
	}

	// Cleanup
	os.Remove("post.json")
	os.Remove("results.txt")
	fmt.Println()
}

// Test 3: Real-World Scenario
func testRealWorldScenario(ctx context.Context, apiKey string) {
	fmt.Println("🌍 Testing Real-World Scenario: API Monitoring System...")

	fsTool := tools.NewFileSystemTool()
	httpTool := tools.NewHTTPRequestTool()
	dtTool := tools.NewDateTimeTool()
	mathTool := tools.NewMathTool()

	ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithTools(fsTool, httpTool, dtTool, mathTool).
		WithAutoExecute(true).
		WithMaxToolRounds(15)

	scenario := `
You are an API monitoring system. Perform these tasks:
1. Get the current UTC time
2. Fetch data from https://jsonplaceholder.typicode.com/posts/1
3. Calculate response time statistics if you make 3 requests: [120ms, 150ms, 130ms]
4. Save a monitoring report to monitor_report.txt with:
   - Timestamp
   - API endpoint tested
   - Average response time
   - Status: OK or FAILED
`

	fmt.Printf("  ▶ Running complex monitoring scenario...\n")
	start := time.Now()
	response, err := ai.Ask(ctx, scenario)
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("  ❌ FAILED (%v) - %v\n", duration, err)
	} else {
		fmt.Printf("  ✅ PASSED (%v)\n", duration)
		fmt.Printf("  Response: %s\n", response)

		// Check if report file was created
		if _, err := os.Stat("monitor_report.txt"); err == nil {
			content, _ := os.ReadFile("monitor_report.txt")
			fmt.Printf("\n  📄 Generated Report:\n%s\n", string(content))
			os.Remove("monitor_report.txt")
		}
	}
	fmt.Println()
}

// Test 4: Error Handling
func testErrorHandling(ctx context.Context, apiKey string) {
	fmt.Println("⚠️  Testing Error Handling...")

	fsTool := tools.NewFileSystemTool()
	httpTool := tools.NewHTTPRequestTool()
	mathTool := tools.NewMathTool()

	ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithTools(fsTool, httpTool, mathTool).
		WithAutoExecute(true)

	testCases := []struct {
		name          string
		query         string
		expectError   bool
		errorExpected string
	}{
		{"Invalid file read", "Read a file that doesn't exist: nonexistent_file_xyz.txt", true, "should handle gracefully"},
		{"Invalid URL", "Fetch from invalid URL: not-a-valid-url", true, "should reject"},
		{"Division by zero", "Calculate: 10 / 0", true, "should handle math error"},
		{"Invalid equation", "Solve: this is not an equation", true, "should reject malformed input"},
	}

	for _, tc := range testCases {
		fmt.Printf("  ▶ %s: ", tc.name)
		_, err := ai.Ask(ctx, tc.query)

		if tc.expectError {
			if err != nil {
				fmt.Printf("✅ PASSED - Error handled correctly\n")
			} else {
				fmt.Printf("⚠️  WARNING - Expected error but got none\n")
			}
		} else {
			if err != nil {
				fmt.Printf("❌ FAILED - Unexpected error: %v\n", err)
			} else {
				fmt.Printf("✅ PASSED\n")
			}
		}
	}
	fmt.Println()
}

// Test 5: Performance Benchmarks
func testPerformance(ctx context.Context, apiKey string) {
	fmt.Println("⚡ Testing Performance...")

	mathTool := tools.NewMathTool()
	ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithTool(mathTool).
		WithAutoExecute(true)

	// Simple expression benchmark
	fmt.Printf("  ▶ Math expression (10 iterations): ")
	var totalDuration time.Duration
	successCount := 0

	for i := 0; i < 10; i++ {
		start := time.Now()
		_, err := ai.Ask(ctx, fmt.Sprintf("Calculate: %d * 2 + 5", i))
		duration := time.Since(start)
		totalDuration += duration

		if err == nil {
			successCount++
		}
	}

	avgDuration := totalDuration / 10
	fmt.Printf("✅ %d/10 succeeded, avg: %v\n", successCount, avgDuration)

	// HTTP request benchmark
	httpTool := tools.NewHTTPRequestTool()
	ai2 := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithTool(httpTool).
		WithAutoExecute(true)

	fmt.Printf("  ▶ HTTP requests (5 iterations): ")
	totalDuration = 0
	successCount = 0

	for i := 0; i < 5; i++ {
		start := time.Now()
		_, err := ai2.Ask(ctx, "Fetch https://jsonplaceholder.typicode.com/todos/1")
		duration := time.Since(start)
		totalDuration += duration

		if err == nil {
			successCount++
		}
	}

	avgDuration = totalDuration / 5
	fmt.Printf("✅ %d/5 succeeded, avg: %v\n", successCount, avgDuration)
}
