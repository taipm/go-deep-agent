# Design Decisions for Builder API (Option B)

This document records key design decisions made during the implementation of the Builder Pattern API.

## Decision Log

---

### Decision 1: Ask() vs AskE() - Error Handling Strategy

**Date:** TBD
**Status:** ✅ Decided

**Context:**
- Go community prefers explicit error handling
- But for rapid prototyping, panicking can be more convenient
- Need to balance both use cases

**Options Considered:**

1. **Only AskE() with error return**
   ```go
   response, err := agent.Ask(ctx, "Hello")
   if err != nil { ... }
   ```
   - ✅ Go idiomatic
   - ❌ Verbose for simple scripts

2. **Only Ask() that panics**
   ```go
   response := agent.Ask(ctx, "Hello") // panics on error
   ```
   - ✅ Clean for scripts
   - ❌ Not Go idiomatic

3. **Both Ask() (panic) and AskE() (error)** ⭐ CHOSEN
   ```go
   // For scripts/prototypes
   response := agent.Ask(ctx, "Hello")
   
   // For production
   response, err := agent.AskE(ctx, "Hello")
   if err != nil { ... }
   ```
   - ✅ Flexibility
   - ✅ Clear naming convention (E suffix for error)
   - ✅ Users choose what they need

**Decision:** Provide both `Ask()` and `AskE()`

**Reasoning:**
- Follows precedent from stdlib (e.g., `regexp.MustCompile` vs `regexp.Compile`)
- Serves both rapid prototyping and production use cases
- Clear naming makes intention obvious

---

### Decision 2: Return String vs Struct from Ask()

**Date:** TBD
**Status:** ✅ Decided

**Context:**
- Most users just want the text response
- Some need metadata (usage, finish reason, etc.)
- Need balance between simplicity and power

**Options Considered:**

1. **Return string directly** ⭐ CHOSEN for Ask()
   ```go
   response := agent.Ask(ctx, "Hello")
   fmt.Println(response) // Just the text!
   ```
   - ✅ Super simple for common case (90%)
   - ✅ Natural, readable code
   - ❌ No access to metadata

2. **Return Result struct**
   ```go
   result := agent.Ask(ctx, "Hello")
   fmt.Println(result.Content)
   ```
   - ✅ Access to metadata
   - ❌ Extra .Content everywhere

3. **Both: Ask() returns string, GetLastCompletion() for metadata** ⭐ BEST
   ```go
   response := agent.Ask(ctx, "Hello")
   
   // If need metadata
   usage := agent.GetUsage()
   completion := agent.GetLastCompletion()
   ```
   - ✅ Simple for common case
   - ✅ Power when needed
   - ✅ No .Content noise

**Decision:** Ask() returns string, provide GetUsage() and GetLastCompletion()

**Reasoning:**
- Optimizes for the 90% use case (just want text)
- Power users can access metadata when needed
- Follows principle of progressive disclosure

---

### Decision 3: Auto-Memory vs Manual History

**Date:** TBD
**Status:** ✅ Decided

**Context:**
- Managing conversation history is tedious
- But some users want full control
- Need both automatic and manual approaches

**Options Considered:**

1. **Always auto-remember**
   ```go
   chat := agent.New(model)
   chat.Ask(ctx, "Q1") // Auto remembers
   chat.Ask(ctx, "Q2") // Knows Q1
   ```
   - ✅ Super convenient
   - ❌ No control
   - ❌ Surprising behavior

2. **Always manual**
   ```go
   chat := agent.New(model)
   chat.WithHistory(messages)
   ```
   - ✅ Full control
   - ❌ Tedious

3. **Opt-in with WithMemory()** ⭐ CHOSEN
   ```go
   // Manual (default)
   agent.New(model).Ask(ctx, "Q1")
   
   // Auto-memory (opt-in)
   chat := agent.New(model).WithMemory()
   chat.Ask(ctx, "Q1") // Remembers
   chat.Ask(ctx, "Q2") // Has context
   ```
   - ✅ Explicit opt-in
   - ✅ No surprises
   - ✅ Flexible

**Decision:** Opt-in auto-memory with `WithMemory()`

**Reasoning:**
- Principle of least surprise (no automatic behavior)
- Clear and explicit when memory is enabled
- Easy to implement: just a flag and append logic

---

### Decision 4: Message Types - Own vs OpenAI

**Date:** TBD
**Status:** ✅ Decided

**Context:**
- Using openai types requires importing openai-go
- Want to hide implementation details
- Need simple API

**Options Considered:**

1. **Use openai types directly**
   ```go
   import "github.com/openai/openai-go/v3"
   
   Messages: []openai.ChatCompletionMessageParamUnion{
       openai.SystemMessage("helpful"),
   }
   ```
   - ❌ Leaky abstraction
   - ❌ Users must import openai

2. **Own simple types** ⭐ CHOSEN
   ```go
   // No openai import needed!
   type Message struct {
       Role    string
       Content string
   }
   
   agent.System("helpful")
   agent.User("question")
   ```
   - ✅ Clean API
   - ✅ No dependencies exposed
   - ✅ Convert internally

**Decision:** Define own Message types and helpers

**Reasoning:**
- Users should not need to import openai-go
- Simpler, cleaner API
- Easy to extend (add fields) without breaking changes
- Internal conversion is trivial

---

### Decision 5: Tool Calling - Auto-Execute vs Return

**Date:** TBD
**Status:** ⚠️ Under Discussion

**Context:**
- Tool calling often needs iteration (call tool → add result → ask again)
- Should we handle this automatically or let users control?

**Options Considered:**

1. **Auto-execute with callback** ⭐ PREFERRED
   ```go
   agent.WithTool("get_weather", "Get weather", params).
       OnToolCall(func(name string, args map[string]any) string {
           // Execute tool
           return result
       }).
       Ask(ctx, "Weather in Hanoi?")
   // Returns final answer after tool execution!
   ```
   - ✅ Automatic loop
   - ✅ Simple for users
   - ❌ Less control

2. **Return tool calls, manual execution**
   ```go
   result := agent.Ask(ctx, "Weather?")
   if result.ToolCalls != nil {
       // User executes tools manually
       // User calls Ask again with tool results
   }
   ```
   - ✅ Full control
   - ❌ Complex

3. **Both: default auto, opt-out available**
   ```go
   // Auto (default)
   agent.OnToolCall(handler).Ask(ctx, msg)
   
   // Manual (advanced)
   result := agent.WithManualTools().Ask(ctx, msg)
   ```

**Recommendation:** Start with Option 1 (auto), add manual mode later if needed

---

### Decision 6: Streaming - Callback vs Channel

**Date:** TBD
**Status:** ✅ Decided

**Context:**
- Two ways to stream: callbacks vs Go channels
- Need to choose one (or both?)

**Options Considered:**

1. **Callback function** ⭐ CHOSEN
   ```go
   agent.OnStream(func(chunk string) {
       fmt.Print(chunk)
   }).Ask(ctx, "Story")
   ```
   - ✅ Simple
   - ✅ Low overhead
   - ✅ Matches openai-go style

2. **Channel**
   ```go
   ch := agent.StreamCh(ctx, "Story")
   for chunk := range ch {
       fmt.Print(chunk)
   }
   ```
   - ✅ Go idiomatic
   - ❌ More complex
   - ❌ Need goroutine management

**Decision:** Use callback functions

**Reasoning:**
- Simpler for most use cases
- Matches openai-go's streaming API
- Less overhead than channels
- Can add channel-based API later if needed

---

### Decision 7: JSON Schema - Helper Methods

**Date:** TBD
**Status:** ✅ Decided

**Context:**
- JSON Schema can be verbose
- Need balance between type-safety and simplicity

**Options Considered:**

1. **Raw map[string]any** ⭐ CHOSEN
   ```go
   agent.WithJSONSchema("person", map[string]any{
       "type": "object",
       "properties": map[string]any{
           "name": map[string]string{"type": "string"},
           "age": map[string]string{"type": "integer"},
       },
   })
   ```
   - ✅ Flexible
   - ✅ Matches JSON Schema spec
   - ❌ Not type-safe

2. **Typed builder**
   ```go
   schema := agent.Schema().
       Object().
       Field("name", agent.String()).
       Field("age", agent.Integer()).
       Build()
   ```
   - ✅ Type-safe
   - ❌ Complex API
   - ❌ Over-engineering for now

**Decision:** Use map[string]any, consider typed builder for v2.1

**Reasoning:**
- JSON Schema is inherently dynamic
- map[string]any is familiar to Go devs
- Can add type-safe builder later without breaking changes

---

### Decision 8: Error Types - Custom vs Wrapped

**Date:** TBD
**Status:** ✅ Decided

**Context:**
- Need good error handling
- Should we define custom error types?

**Options Considered:**

1. **Just wrap errors**
   ```go
   return fmt.Errorf("chat failed: %w", err)
   ```
   - ✅ Simple
   - ❌ Hard to handle specific errors

2. **Custom error types** ⭐ CHOSEN
   ```go
   type ErrRateLimit struct {
       RetryAfter time.Duration
   }
   
   if errors.As(err, &ErrRateLimit{}) {
       // Handle rate limit
   }
   ```
   - ✅ Type-safe error handling
   - ✅ Can attach metadata
   - ✅ Better UX

**Decision:** Define custom error types for common cases

**Errors to define:**
- `ErrAPIKey` - missing/invalid API key
- `ErrRateLimit` - rate limit with retry info
- `ErrTimeout` - request timeout
- `ErrRefusal` - content refused by model
- `ErrInvalidResponse` - malformed API response

---

### Decision 9: Context Window Management

**Date:** TBD
**Status:** ⚠️ Future Decision

**Context:**
- Conversations can exceed context window
- Need strategy to handle this

**Options:**

1. **Manual (user handles)**
2. **Auto-truncate oldest messages**
3. **Summarize old messages**
4. **Sliding window with summary**

**Decision:** Defer to Phase 7, start with manual

---

### Decision 10: Retry Strategy

**Date:** TBD
**Status:** ⚠️ Future Decision

**Context:**
- API calls can fail transiently
- Need retry strategy

**Options:**

1. **No retry (user handles)**
2. **Simple retry with exponential backoff**
3. **Smart retry (only transient errors)**

**Decision:** Defer to Phase 8

---

## Principles

Throughout design, follow these principles:

1. **Simplicity First** - Optimize for the common case
2. **Progressive Disclosure** - Hide complexity, expose when needed
3. **No Surprises** - Explicit is better than implicit
4. **Go Idiomatic** - Follow Go community conventions
5. **Extensibility** - Easy to add features without breaking changes

---

## Open Questions

1. Should `Ask()` auto-add system message from `WithSystem()`?
   - Currently: Yes (makes sense for memory mode)
   
2. Max retry attempts default?
   - Recommendation: 3
   
3. Should streaming return the full text at end?
   - Currently: Yes (useful for logging)
   
4. Context window limit detection?
   - Defer to Phase 7

5. Support for function calling with parallel tool calls?
   - Defer to Phase 4 implementation

---

## Changes Log

Track any design changes here:

| Date | Decision # | Change | Reason |
|------|-----------|--------|--------|
| TBD | - | Initial decisions | - |

---

## Review Checklist

Before finalizing any decision:

- [ ] Does it simplify the API?
- [ ] Is it Go idiomatic?
- [ ] Can it be extended later?
- [ ] Does it surprise users?
- [ ] Is it documented?
- [ ] Are there examples?
