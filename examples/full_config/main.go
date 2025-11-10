package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/taipm/go-deep-agent/agent"
)

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable not set")
	}

	ctx := context.Background()

	fmt.Println("🔧 Full Configuration Demo - Phase 3 (v0.6.4)")
	fmt.Println("=" + string(make([]byte, 60)) + "=\n")

	// ========================================
	// Demo 1: Persona Only (Phase 2)
	// ========================================
	fmt.Println("📋 Demo 1: Persona Only (Phase 2 approach)")
	fmt.Println("-" + string(make([]byte, 60)) + "-")

	persona, err := agent.LoadPersona("personas/customer_support.yaml")
	if err != nil {
		log.Fatalf("Failed to load persona: %v", err)
	}

	agentPersonaOnly := agent.NewOpenAI("", apiKey).
		WithPersona(persona)

	fmt.Printf("✅ Loaded persona: %s\n", persona.Name)
	fmt.Printf("   Model: %s\n", persona.GetModel())
	fmt.Printf("   Temperature: %.1f\n", persona.GetTemperature())
	fmt.Printf("   Role: %s\n\n", persona.Role)

	response1, err := agentPersonaOnly.Ask(ctx, "I can't log into my account!")
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("🤖 Response: %s\n\n", truncate(response1, 200))
	}

	// ========================================
	// Demo 2: Settings Only (Technical config)
	// ========================================
	fmt.Println("\n📋 Demo 2: Settings Only (Technical configuration)")
	fmt.Println("-" + string(make([]byte, 60)) + "-")

	settings, err := agent.LoadSettings("configs/settings_production.yaml")
	if err != nil {
		log.Fatalf("Failed to load settings: %v", err)
	}

	agentSettingsOnly := agent.NewOpenAI("", apiKey).
		WithSystem("You are a helpful assistant.").
		WithSettings(settings)

	fmt.Printf("✅ Loaded settings:\n")
	fmt.Printf("   Model: %s\n", settings.Model)
	fmt.Printf("   Temperature: %.1f\n", settings.Temperature)
	fmt.Printf("   Max Tokens: %d\n", settings.MaxTokens)
	fmt.Printf("   Memory Capacity: %d\n", settings.Memory.WorkingCapacity)
	fmt.Printf("   Retry Attempts: %d\n\n", settings.Retry.MaxAttempts)

	response2, err := agentSettingsOnly.Ask(ctx, "What is Go?")
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("🤖 Response: %s\n\n", truncate(response2, 200))
	}

	// ========================================
	// Demo 3: Full Config (Persona + Settings) - NEW!
	// ========================================
	fmt.Println("\n📋 Demo 3: Full Config (Persona + Settings Override) - Phase 3!")
	fmt.Println("-" + string(make([]byte, 60)) + "-")

	fullConfig, err := agent.LoadFullConfig("configs/full_customer_support.yaml")
	if err != nil {
		log.Fatalf("Failed to load full config: %v", err)
	}

	agentFull := agent.NewOpenAI("", apiKey).
		WithFullConfig(fullConfig)

	fmt.Printf("✅ Loaded full configuration:\n")
	if fullConfig.Metadata != nil {
		fmt.Printf("   Name: %s\n", fullConfig.Metadata.Name)
		fmt.Printf("   Environment: %s\n", fullConfig.Metadata.Environment)
	}
	fmt.Printf("   Persona: %s (role: %s)\n", fullConfig.Persona.Name, fullConfig.Persona.Role)
	fmt.Printf("   Persona's model: %s (temp: %.1f)\n",
		fullConfig.Persona.GetModel(),
		fullConfig.Persona.GetTemperature())
	fmt.Printf("   Settings override model: %s (temp: %.1f)\n",
		fullConfig.Settings.Model,
		fullConfig.Settings.Temperature)
	fmt.Printf("   ⚙️  Final model used: %s\n", fullConfig.GetModel())
	fmt.Printf("   ⚙️  Final temperature: %.1f\n\n", fullConfig.GetTemperature())

	response3, err := agentFull.Ask(ctx, "My order hasn't arrived yet!")
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("🤖 Response: %s\n\n", truncate(response3, 200))
	}

	// ========================================
	// Demo 4: Settings Override Persona
	// ========================================
	fmt.Println("\n📋 Demo 4: Settings Override Persona (Flexible override)")
	fmt.Println("-" + string(make([]byte, 60)) + "-")

	// Load persona (has gpt-4o-mini, temp 0.7)
	personaBase, _ := agent.LoadPersona("personas/customer_support.yaml")

	// Load production settings (has gpt-4, temp 0.6)
	prodSettings, _ := agent.LoadSettings("configs/settings_production.yaml")

	// Apply persona first, then override with settings
	agentOverride := agent.NewOpenAI("", apiKey).
		WithPersona(personaBase).
		WithSettings(prodSettings)

	fmt.Printf("✅ Override demonstration:\n")
	fmt.Printf("   Persona model: %s (temp: %.1f)\n",
		personaBase.GetModel(),
		personaBase.GetTemperature())
	fmt.Printf("   Settings model: %s (temp: %.1f)\n",
		prodSettings.Model,
		prodSettings.Temperature)
	fmt.Printf("   ⚙️  Final model: %s (Settings WIN!)\n", prodSettings.Model)
	fmt.Printf("   ⚙️  Final temp: %.1f (Settings WIN!)\n\n", prodSettings.Temperature)

	response4, err := agentOverride.Ask(ctx, "How do I reset my password?")
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("🤖 Response: %s\n\n", truncate(response4, 200))
	}

	// ========================================
	// Demo 5: Code Override Everything
	// ========================================
	fmt.Println("\n📋 Demo 5: Code Overrides Everything (Ultimate flexibility)")
	fmt.Println("-" + string(make([]byte, 60)) + "-")

	config2, _ := agent.LoadFullConfig("configs/full_customer_support.yaml")

	agentCodeOverride := agent.NewOpenAI("", apiKey).
		WithFullConfig(config2).
		WithTemperature(0.9). // Code override!
		WithMaxTokens(500)    // Code override!

	fmt.Printf("✅ Override priority demonstration:\n")
	fmt.Printf("   Config temperature: %.1f\n", config2.GetTemperature())
	fmt.Printf("   Code override temperature: 0.9\n")
	fmt.Printf("   ⚙️  Priority: Code > Settings > Persona > Defaults\n\n")

	response5, err := agentCodeOverride.Ask(ctx, "Tell me about your refund policy")
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("🤖 Response: %s\n\n", truncate(response5, 200))
	}

	// ========================================
	// Demo 6: Development vs Production Settings
	// ========================================
	fmt.Println("\n📋 Demo 6: Development vs Production Settings")
	fmt.Println("-" + string(make([]byte, 60)) + "-")

	devSettings, _ := agent.LoadSettings("configs/settings_development.yaml")
	prodSettings2, _ := agent.LoadSettings("configs/settings_production.yaml")

	fmt.Printf("🔧 Development Settings:\n")
	fmt.Printf("   Model: %s (cheaper, faster)\n", devSettings.Model)
	fmt.Printf("   Max Tokens: %d (lower limit)\n", devSettings.MaxTokens)
	fmt.Printf("   Retry: %d attempts (fail fast)\n", devSettings.Retry.MaxAttempts)
	fmt.Printf("   Parallel Tools: %v (easier debugging)\n\n", devSettings.Tools.ParallelExecution)

	fmt.Printf("🚀 Production Settings:\n")
	fmt.Printf("   Model: %s (more capable)\n", prodSettings2.Model)
	fmt.Printf("   Max Tokens: %d (higher limit)\n", prodSettings2.MaxTokens)
	fmt.Printf("   Retry: %d attempts (more reliable)\n", prodSettings2.Retry.MaxAttempts)
	fmt.Printf("   Parallel Tools: %v (better performance)\n\n", prodSettings2.Tools.ParallelExecution)

	// ========================================
	// Summary
	// ========================================
	fmt.Println("\n" + string(make([]byte, 60)) + "")
	fmt.Println("📊 Summary - Configuration Options")
	fmt.Println(string(make([]byte, 60)) + "\n")

	fmt.Println("1. ✅ Persona Only:")
	fmt.Println("   - Simple behavior-first approach")
	fmt.Println("   - Perfect for non-technical users")
	fmt.Println("   - Use: LoadPersona() + WithPersona()\n")

	fmt.Println("2. ✅ Settings Only:")
	fmt.Println("   - Technical configuration without behavior")
	fmt.Println("   - Perfect for engineers fine-tuning")
	fmt.Println("   - Use: LoadSettings() + WithSettings()\n")

	fmt.Println("3. ⭐ Full Config (NEW!):")
	fmt.Println("   - Best of both worlds")
	fmt.Println("   - Persona + Settings in one file")
	fmt.Println("   - Use: LoadFullConfig() + WithFullConfig()\n")

	fmt.Println("4. 🎯 Flexible Override:")
	fmt.Println("   - Mix and match as needed")
	fmt.Println("   - Persona + separate Settings files")
	fmt.Println("   - Use: WithPersona() + WithSettings()\n")

	fmt.Println("5. 🔧 Code Override:")
	fmt.Println("   - Code always wins")
	fmt.Println("   - Dynamic runtime configuration")
	fmt.Println("   - Use: WithFullConfig() + With*()\n")

	fmt.Println("\n🎉 Phase 3 Complete! Maximum flexibility achieved!")
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
