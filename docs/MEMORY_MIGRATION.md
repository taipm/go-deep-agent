# Memory System Migration Guide

Guide for migrating from v0.5.x (simple FIFO memory) to current version (hierarchical memory with episodic storage).

## Overview

The memory system has been significantly enhanced with **hierarchical memory architecture**:

- **v0.5.x**: Simple FIFO buffer (First In, First Out)
- **Current**: 3-tier hierarchical system (Working → Episodic → Semantic)

## What Changed

### Architecture Evolution

**Before (v0.5.x):**
```
Messages → FIFO Buffer (fixed size) → Compression when full
```

**After (Current):**
```
Messages → Working Memory (FIFO, hot storage)
         ↓ (importance >= threshold)
         → Episodic Memory (vector-based, long-term)
         → Semantic Memory (facts/knowledge)
```

### Key Differences

| Feature | v0.5.x | Current |
|---------|--------|---------|
| Memory Type | Single FIFO buffer | 3-tier hierarchy |
| Important Messages | Treated same as all | Auto-stored in episodic |
| Recall | Recent messages only | Semantic search + recency |
| Compression | Simple summarization | Smart tier-based |
| Importance Scoring | None | Automatic with weights |
| Fact Storage | None | Semantic memory support |

## Breaking Changes

### ⚠️ None! (100% Backward Compatible)

The new memory system is **fully backward compatible**. Your existing code will work without changes:

```go
// v0.5.x code - still works!
builder := agent.NewOpenAI("gpt-4o-mini", apiKey)
response, err := builder.Ask(ctx, "Hello")
```

Default behavior:
- Working memory capacity: 10 messages
- Episodic memory: **enabled** by default (threshold: 0.5)
- Importance scoring: **enabled** by default
- Semantic memory: **disabled** by default

## Migration Paths

### Option 1: Keep v0.5.x Behavior (No Migration Needed)

If you want the exact v0.5.x behavior (simple FIFO, no episodic):

```go
builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
    DisableMemory()  // Disable hierarchical memory (use simple FIFO)
```

### Option 2: Use Defaults (Automatic Upgrade)

Just use the new version as-is - you get episodic memory automatically:

```go
// Your existing code - automatically gets episodic memory!
builder := agent.NewOpenAI("gpt-4o-mini", apiKey)

// Important messages are now automatically stored in episodic
response, err := builder.Ask(ctx, "Remember: my birthday is May 5th")
```

### Option 3: Custom Configuration (Recommended)

Take advantage of new features with custom config:

```go
// Configure episodic memory threshold
builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithWorkingMemorySize(20).        // Increase working capacity
    WithEpisodicMemory(0.7)            // Only store high-importance messages

// Or use full config
config := memory.DefaultMemoryConfig()
config.EpisodicThreshold = 0.8        // Higher threshold
config.WorkingCapacity = 50           // Larger working memory

builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithHierarchicalMemory(config)
```

## Common Migration Scenarios

### Scenario 1: Basic Chatbot (v0.5.x → Current)

**Before:**
```go
builder := agent.NewOpenAI("gpt-4o-mini", apiKey)
response, _ := builder.Ask(ctx, "Hello")
```

**After (no changes required, but can enhance):**
```go
// Option A: Keep as-is (gets episodic automatically)
builder := agent.NewOpenAI("gpt-4o-mini", apiKey)

// Option B: Customize for your use case
builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithEpisodicMemory(0.7).          // Store important convos
    WithWorkingMemorySize(15)         // Remember last 15 messages
```

### Scenario 2: Long Conversations

**Before:**
```go
// Had to manually manage conversation history
// Messages would be lost after buffer filled
builder := agent.NewOpenAI("gpt-4o-mini", apiKey)
```

**After:**
```go
// Important messages automatically preserved!
builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithEpisodicMemory(0.6).          // Lower threshold = more memories
    WithWorkingMemorySize(30)         // Larger working buffer

// Access memory stats
mem := builder.GetMemory()
stats := mem.Stats(ctx)
fmt.Printf("Stored %d important messages\n", stats.EpisodicSize)
```

### Scenario 3: Custom Importance Logic

**Before:**
```go
// No built-in importance scoring
// Had to implement manually
```

**After:**
```go
// Customize what's considered "important"
weights := memory.DefaultImportanceWeights()
weights.ExplicitRemember = 2.0       // Double weight for "remember this"
weights.PersonalInfo = 1.5           // Higher weight for personal info
weights.QuestionAnswer = 0.3         // Lower weight for Q&A

builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithImportanceWeights(weights).
    WithEpisodicMemory(0.8)           // High threshold
```

### Scenario 4: Knowledge/Fact Storage

**Before:**
```go
// No built-in fact storage
// Had to use external database
```

**After:**
```go
// Use semantic memory for facts
builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithSemanticMemory().
    WithEpisodicMemory(0.6)

// Store facts
mem := builder.GetMemory()
mem.StoreFact(ctx, memory.Fact{
    Content:  "User prefers Python over JavaScript",
    Category: "preferences",
})
```

## New Features You Can Use

### 1. Memory Statistics

```go
mem := builder.GetMemory()
stats := mem.Stats(ctx)

fmt.Printf("Working memory: %d messages\n", stats.WorkingSize)
fmt.Printf("Episodic memory: %d important messages\n", stats.EpisodicSize)
fmt.Printf("Average importance: %.2f\n", stats.AverageImportance)
fmt.Printf("Oldest memory: %s\n", stats.EpisodicOldest)
fmt.Printf("Newest memory: %s\n", stats.EpisodicNewest)
```

### 2. Custom Recall

```go
// Recall specific memories
opts := memory.DefaultRecallOptions()
opts.EpisodicTopK = 5              // Get 5 most relevant episodic memories
opts.MinImportance = 0.7           // Only high-importance

messages, err := mem.Recall(ctx, "birthday", opts)
```

### 3. Memory Inspection

```go
// Check configuration
config := mem.GetConfig()
fmt.Printf("Episodic enabled: %v\n", config.EpisodicEnabled)
fmt.Printf("Threshold: %.2f\n", config.EpisodicThreshold)

// Update configuration
config.EpisodicThreshold = 0.8
mem.SetConfig(config)
```

## Performance Considerations

### Memory Usage

**v0.5.x:**
- Fixed buffer size (default: 10 messages)
- ~1KB per message
- Total: ~10KB

**Current:**
- Working: ~10 messages (~10KB)
- Episodic: Variable, but typically ~100-1000 messages
- Estimated: ~100KB - 1MB for typical usage

### Speed

- Working memory: Same as v0.5.x (instant)
- Episodic storage: ~1-5ms per message
- Episodic recall: ~10-50ms for semantic search
- **No noticeable impact** on chat response times

### Optimization Tips

```go
// For high-volume applications
config := memory.DefaultMemoryConfig()
config.WorkingCapacity = 50          // Larger buffer
config.EpisodicThreshold = 0.8       // Store less
config.EpisodicMaxSize = 1000        // Limit episodic size

builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithHierarchicalMemory(config)
```

## Troubleshooting

### Issue: Too many messages stored in episodic

**Solution:** Increase threshold
```go
builder.WithEpisodicMemory(0.9)  // Only very important messages
```

### Issue: Important messages not being stored

**Solution:** Lower threshold or adjust weights
```go
builder.WithEpisodicMemory(0.5)  // Lower threshold

// Or customize weights
weights := memory.DefaultImportanceWeights()
weights.ExplicitRemember = 2.0
builder.WithImportanceWeights(weights)
```

### Issue: Memory using too much RAM

**Solution:** Set max size limits
```go
config := memory.DefaultMemoryConfig()
config.EpisodicMaxSize = 500       // Limit to 500 messages
builder.WithHierarchicalMemory(config)
```

### Issue: Want v0.5.x behavior exactly

**Solution:** Disable hierarchical memory
```go
builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
    DisableMemory()  // Use simple FIFO like v0.5.x
```

## Testing Your Migration

### 1. Verify Episodic Storage

```go
mem := builder.GetMemory()

// Send important message
builder.Ask(ctx, "Remember: my favorite color is blue")

// Check stats
stats := mem.Stats(ctx)
if stats.EpisodicSize > 0 {
    fmt.Println("✅ Episodic memory working!")
} else {
    fmt.Println("❌ Check threshold and weights")
}
```

### 2. Verify Recall

```go
// Store something
builder.Ask(ctx, "Remember: I'm allergic to peanuts")

// Later, recall
builder.Ask(ctx, "What did I tell you about my allergies?")
// Should recall the peanut allergy from episodic memory
```

### 3. Monitor Performance

```go
start := time.Now()
response, err := builder.Ask(ctx, "Hello")
duration := time.Since(start)

fmt.Printf("Response time: %v\n", duration)
// Should be similar to v0.5.x (within 10-20ms)
```

## API Reference

### New Builder Methods

- `WithEpisodicMemory(threshold float64)` - Enable episodic memory
- `WithImportanceWeights(weights)` - Customize importance calculation
- `WithWorkingMemorySize(size int)` - Set working memory capacity
- `WithSemanticMemory()` - Enable fact storage
- `WithHierarchicalMemory(config)` - Full configuration
- `GetMemory()` - Access memory system directly

### New Memory Methods

- `Stats(ctx) MemoryStats` - Get memory statistics
- `Recall(ctx, query, opts) ([]Message, error)` - Custom recall
- `GetConfig() MemoryConfig` - Get current configuration
- `SetConfig(config) error` - Update configuration

## Examples

See complete examples in:
- `examples/builder_memory_integration.go` - Configuration examples
- `examples/e2e_integration.go` - Full integration test
- `agent/builder_memory_test.go` - Unit tests

## Need Help?

- Check `docs/MEMORY_ARCHITECTURE.md` for architecture details
- Run `examples/e2e_integration.go` to see it in action
- See `agent/memory/` for source code and tests

## Summary

✅ **Backward compatible** - existing code works without changes
✅ **Automatic upgrades** - get episodic memory by default
✅ **Easy customization** - new Builder methods for configuration
✅ **No performance impact** - memory operations are fast
✅ **Optional features** - use as much or as little as you need

**Recommended migration:**
1. Test with defaults (automatic episodic memory)
2. Monitor memory stats
3. Adjust threshold/weights based on your use case
4. Enable semantic memory if storing facts

Happy migrating! 🚀
