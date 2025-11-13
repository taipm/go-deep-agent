package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/taipm/go-deep-agent/agent"
)

// Session Persistence - Basic Example
//
// This example demonstrates the new WithSessionID() feature (v0.8.0+)
// which enables automatic conversation persistence.
//
// Features demonstrated:
//  1. Auto-save conversation after each message
//  2. Auto-load previous conversation on restart
//  3. Zero-configuration (uses default FileBackend)
//
// Run this example multiple times to see session persistence in action!

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	ctx := context.Background()

	// ========================================
	// Example 1: Simple Session Persistence
	// ========================================
	fmt.Println("=== Example 1: Simple Session Persistence ===\n")

	// Create agent with session ID
	// This will:
	//  1. Check if session "demo-user-alice" exists
	//  2. If yes, load previous conversation
	//  3. Auto-save after each Ask()
	agent1 := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithMemory().
		WithSessionID("demo-user-alice") // ← Magic happens here!

	// First conversation (or continuation if restarted)
	response1, err := agent1.Ask(ctx, "My name is Alice and I'm from Vietnam")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("AI: %s\n\n", response1)

	// Second message - AI remembers context
	response2, err := agent1.Ask(ctx, "What's my name and where am I from?")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("AI: %s\n\n", response2)

	// Session is automatically saved!
	fmt.Println("✅ Session automatically saved to: ~/.go-deep-agent/sessions/demo-user-alice.json")
	fmt.Println("💡 Tip: Run this example again - the conversation will continue!")
	fmt.Println()

	// ========================================
	// Example 2: Multiple Sessions
	// ========================================
	fmt.Println("=== Example 2: Multiple Sessions (Per-User) ===\n")

	// User Bob's session (separate from Alice)
	agent2 := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithMemory().
		WithSessionID("demo-user-bob")

	response3, err := agent2.Ask(ctx, "My name is Bob and I love coding in Go")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Bob's Agent: %s\n\n", response3)

	// User Charlie's session (separate from Alice and Bob)
	agent3 := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithMemory().
		WithSessionID("demo-user-charlie")

	response4, err := agent3.Ask(ctx, "My name is Charlie and I'm learning AI")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Charlie's Agent: %s\n\n", response4)

	fmt.Println("✅ Three separate sessions created:")
	fmt.Println("   - ~/.go-deep-agent/sessions/demo-user-alice.json")
	fmt.Println("   - ~/.go-deep-agent/sessions/demo-user-bob.json")
	fmt.Println("   - ~/.go-deep-agent/sessions/demo-user-charlie.json")
	fmt.Println()

	// ========================================
	// Example 3: Session Management
	// ========================================
	fmt.Println("=== Example 3: Session Management ===\n")

	// List all sessions
	sessions, err := agent1.ListSessions(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Available sessions: %d\n", len(sessions))
	for i, sessionID := range sessions {
		fmt.Printf("  %d. %s\n", i+1, sessionID)
	}
	fmt.Println()

	// Get current session ID
	currentSession := agent1.GetSessionID()
	fmt.Printf("Current session: %s\n", currentSession)

	// Get conversation history
	history := agent1.GetHistory()
	fmt.Printf("Messages in current session: %d\n", len(history))
	fmt.Println()

	// ========================================
	// Example 4: Manual Save/Load Control
	// ========================================
	fmt.Println("=== Example 4: Manual Save/Load (Advanced) ===\n")

	// Disable auto-save for manual control
	agent4 := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithMemory().
		WithSessionID("demo-user-manual").
		WithAutoSave(false) // ← Manual mode

	response5, err := agent4.Ask(ctx, "Message 1")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("AI: %s\n", response5)

	response6, err := agent4.Ask(ctx, "Message 2")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("AI: %s\n\n", response6)

	// Explicit save
	if err := agent4.SaveSession(ctx); err != nil {
		log.Fatal(err)
	}
	fmt.Println("✅ Session manually saved")
	fmt.Println()

	// ========================================
	// Example 5: Session Cleanup
	// ========================================
	fmt.Println("=== Example 5: Session Cleanup ===\n")

	// Create a temporary session
	tempAgent := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithMemory().
		WithSessionID("demo-temp-session")

	_, err = tempAgent.Ask(ctx, "This is a temporary conversation")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Created temporary session: demo-temp-session")

	// Delete the session
	if err := tempAgent.DeleteSession(ctx); err != nil {
		log.Fatal(err)
	}

	fmt.Println("✅ Temporary session deleted")
	fmt.Println()

	// ========================================
	// Summary
	// ========================================
	fmt.Println("=== Summary ===\n")
	fmt.Println("🎯 Session Persistence Benefits:")
	fmt.Println("   ✅ Conversations survive app restarts")
	fmt.Println("   ✅ Per-user session management")
	fmt.Println("   ✅ Automatic save/load (zero code)")
	fmt.Println("   ✅ File-based storage (no external deps)")
	fmt.Println()
	fmt.Println("📂 Sessions stored in:")
	fmt.Println("   ~/.go-deep-agent/sessions/")
	fmt.Println()
	fmt.Println("🔗 Next steps:")
	fmt.Println("   - Run this example again to see persistence in action")
	fmt.Println("   - Check session files: ls ~/.go-deep-agent/sessions/")
	fmt.Println("   - Try session_persistence_advanced.go for more features")
}
