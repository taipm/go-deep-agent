# TODO: Option B - Full Rewrite with Builder Pattern

Roadmap for implementing fluent Builder API that maximizes openai-go capabilities.

## 🎯 Goals

- ✅ Fluent, natural API (method chaining)
- ✅ Tận dụng 100% openai-go features
- ✅ No need to import openai-go in user code
- ✅ Auto conversation memory
- ✅ Simple for beginners, powerful for experts
- ✅ Clean, maintainable codebase

## 📊 Current Status

**Progress:** 8/12 phases complete ✅ **PRODUCTION-READY WITH ROBUST ERROR HANDLING**

**Completed Phases:**
- ✅ Phase 1: Core Builder (12 tests)
- ✅ Phase 2: Advanced Parameters (9 tests)
- ✅ Phase 3: Full Streaming (3 tests)
- ✅ Phase 4: Tool Calling (19 tests)
- ✅ Phase 5: JSON Schema (3 tests)
- ✅ Phase 6: Testing & Documentation (55 tests, 39.2% coverage)
- ✅ Phase 7: Conversation Management (7 tests, 6 examples)
- ✅ Phase 8: Error Handling & Recovery (14 tests, 6 examples)

**Quality Metrics:**
- 76 tests, all passing (0 failures) ⬆️ +14 from Phase 8
- 50.9% overall coverage ⬆️ +9.2% from Phase 8 🎉
- 100% coverage on all configuration methods
- 8 example files with 34+ working examples
- Comprehensive documentation (BUILDER_API.md, TEST_COVERAGE.md, JSON_SCHEMA.md)
- Production-ready error handling (timeout, retry, exponential backoff)
- Integration tests verified with OpenAI and Ollama

**Remaining:** Phases 9-12 (examples polish, benchmarks, advanced features, v2.0.0 release)

---

## Phase 1: Core Builder Implementation ✅ COMPLETE

### 1.1 Message Types & Helpers ✅
- [x] Create `agent/message.go`
  - [x] `Message` struct (Role, Content)
  - [x] `System(content)` helper
  - [x] `User(content)` helper
  - [x] `Assistant(content)` helper
  - [x] `convertMessages()` to openai types

### 1.2 Builder Core ✅
- [x] Create `agent/builder.go`
  - [x] `Builder` struct with all fields
    - [x] Core: model, provider, apiKey, baseURL
    - [x] Conversation: system, messages, memory flag
    - [x] Advanced params: temperature, topP, maxTokens, etc.
    - [x] Streaming: callbacks
    - [x] Tools: tool definitions, handlers
    - [x] Response format: JSON Schema support
    - [x] Internal: client, lastCompletion
  - [x] Constructor methods
    - [x] `New(provider, model) *Builder`
    - [x] `NewOpenAI(model, apiKey string) *Builder`
    - [x] `NewOllama(model string) *Builder`

### 1.3 Configuration Methods ✅
- [x] Basic config
  - [x] `WithAPIKey(key string) *Builder`
  - [x] `WithBaseURL(url string) *Builder`
  - [x] `WithSystem(message string) *Builder`
  - [x] `WithMemory(bool) *Builder`
  - [x] `WithMessages([]Message) *Builder`

### 1.4 Execution Methods ✅
- [x] Simple execution
  - [x] `Ask(ctx, message) (string, error)` - returns error
  - [x] `ensureClient()` - lazy client initialization
  - [x] `buildParams()` - convert builder to openai params
  - [x] `executeSyncRaw()` - non-streaming execution
  - [x] `addMessage()` - add to conversation history
  - [x] `buildMessages()` - prepare messages for API

### 1.5 Basic Tests ✅
- [x] Test constructors (New, NewOpenAI, NewOllama)
- [x] Test WithAPIKey, WithBaseURL
- [x] Test WithSystem, WithMemory, WithMessages
- [x] Test buildMessages()
- [x] Test message helpers
- [x] 12 tests passing

---

## Phase 2: Advanced Parameters (Maximize openai-go) ✅ COMPLETE

### 2.1 Generation Parameters ✅
- [x] `WithTemperature(t float64) *Builder`
- [x] `WithTopP(p float64) *Builder`
- [x] `WithMaxTokens(max int64) *Builder`
- [x] `WithPresencePenalty(p float64) *Builder`
- [x] `WithFrequencyPenalty(p float64) *Builder`
- [x] `WithSeed(seed int64) *Builder`

### 2.2 Analysis Features ✅
- [x] `WithLogprobs(bool) *Builder`
- [x] `WithTopLogprobs(n int64) *Builder`
- [x] `WithMultipleChoices(n int64) *Builder`
- [x] `AskMultiple(ctx, message) ([]string, error)` - implemented

### 2.3 Tests ✅
- [x] Test all advanced parameters (9 tests)
- [x] Test method chaining
- [x] Test buildParams includes all parameters
- [x] Examples in builder_advanced.go (6 examples working)

---

## Phase 3: Streaming (Full ChatCompletionAccumulator)

## Phase 3: Full Streaming Support ✅ COMPLETE

### 3.1 Streaming Core ✅
- [x] `OnStream(callback func(string)) *Builder`
- [x] `Stream(ctx, message) error` - complete streaming implementation
- [x] `StreamPrint(ctx, message) error` - convenience method for terminal output

### 3.2 Advanced Streaming Callbacks ✅
- [x] Use `JustFinishedContent()` for content streaming
- [x] Use `JustFinishedToolCalls()` for tool streaming
- [x] Support streaming with tool calling (auto-execute in stream)
- [x] Support streaming with multiple choices

### 3.3 Refusal Handling ✅
- [x] `OnRefusal(callback func(string)) *Builder`
- [x] Handle refusal in streaming responses
- [x] Track refusal in message history

### 3.4 Tests ✅
- [x] Test streaming callbacks (OnStream, OnToolCall, OnRefusal - 3 tests)
- [x] Test StreamPrint terminal output
- [x] Examples in builder_stream.go (5 streaming examples working)
- [x] Integration tests with OpenAI verified

---

## Phase 4: Tool Calling (Fixed & Enhanced) ✅ COMPLETE

### 4.1 Tool Definition ✅
- [x] `Tool` struct (Name, Description, Parameters, Handler)
- [x] `AddParameter(name, type, description, required)` method with chaining
- [x] Parameter helpers: StringParam, NumberParam, BoolParam, ArrayParam, EnumParam
- [x] Complex parameter support (nested objects)

### 4.2 Tool Execution ✅
- [x] `OnToolCall(handler func(name, args) string) *Builder`
- [x] `WithTool(tool) *Builder` - single tool registration
- [x] `WithTools(tools...) *Builder` - multiple tools
- [x] `WithAutoExecute(bool) *Builder` - auto tool execution in Ask/Stream
- [x] `WithMaxToolRounds(int) *Builder` - prevent infinite loops (default 5)
- [x] Tool handler support on Tool struct itself

### 4.3 Tests ✅
- [x] Test tool creation (15 tests in tool_test.go, 100% coverage)
- [x] Test parameter types and chaining
- [x] Test tool handler execution
- [x] Test WithTool, WithTools, WithAutoExecute, WithMaxToolRounds (4 tests)
- [x] Examples in builder_tools.go (6 comprehensive examples working)
- [x] Integration tests with OpenAI verified

---

## Phase 5: JSON Schema (Structured Outputs) ✅ COMPLETE

### 5.1 JSON Schema Support ✅
- [x] `responseFormat` field (*ChatCompletionNewParamsResponseFormatUnion)
- [x] `WithJSONSchema(name, description, schema, strict) *Builder`
- [x] Support strict JSON Schema with OfJSONSchema()
- [x] Strict mode validation: all properties required, additionalProperties: false
- [x] Fixed executeSyncRaw() to apply responseFormat via buildParams()

### 5.2 Convenience Methods ✅
- [x] `WithJSONMode() *Builder` - free-form JSON via OfJSONObject()
- [x] `WithResponseFormat(format) *Builder` - custom format support
- [x] Automatic JSON parsing in responses

### 5.3 Tests ✅
- [x] Test WithJSONMode, WithJSONSchema, WithResponseFormat (3 tests)
- [x] 4 comprehensive examples in builder_json_schema.go:
  * JSON Mode with prompt instructions
  * Weather Schema (basic structured output)
  * Data Extraction (person info from text)
  * Nested Structures (book review with nested objects)
- [x] All examples tested successfully with OpenAI API
- [x] Complete JSON_SCHEMA.md documentation guide

---

## Phase 6: Observability & Debugging

## Phase 6: Testing & Documentation ✅ COMPLETE

### 6.1 Comprehensive Test Suite ✅
- [x] agent/tool_test.go: 15 tests for tool functionality (100% coverage of tool.go)
  - Tool creation, parameter definition, handler execution
  - All parameter types: String, Number, Bool, Array, Enum
  - Complex parameters, JSON handling, tool independence
- [x] agent/builder_extensions_test.go: 28 tests for Builder methods
  - All advanced parameters (9 tests)
  - Streaming callbacks (3 tests)
  - Tool configuration (4 tests)
  - JSON Schema (3 tests)
  - Internal methods (4 tests)
- [x] agent/builder_test.go: 12 tests for core Builder (fixed duplicate package)
- [x] **Total: 55 tests, all passing, 39.2% overall coverage**
- [x] **100% coverage on all configuration methods**

### 6.2 Documentation ✅
- [x] **BUILDER_API.md**: Complete Builder API guide (430 lines)
  - Quick start, features, advanced parameters
  - Streaming, tools, JSON Schema
  - 5 complete real-world examples
  - Best practices
- [x] **docs/TEST_COVERAGE.md**: Detailed coverage report (250 lines)
  - Coverage breakdown by module
  - Explains 0% on API-calling methods (requires live endpoints)
  - CI/CD recommendations, quality metrics
- [x] **docs/JSON_SCHEMA.md**: JSON Schema guide (from Phase 5)
  - Complete API reference, examples, best practices

### 6.3 Coverage Analysis ✅
- [x] Generated coverage.out and coverage.html reports
- [x] 100% coverage achieved on: tool.go, message.go, all With* methods
- [x] 0% coverage on API methods (Ask, Stream, etc.) - expected, requires live API
- [x] Quality metrics: 55 tests, 0 failures, production-ready

---

## Phase 7: Conversation Management ✅ COMPLETE

### 7.1 Memory Features ✅
- [x] Auto-remember when `WithMemory()` enabled (already existed)
- [x] `Clear() *Builder` - reset conversation
- [x] `GetHistory() []Message` - get all messages
- [x] `SetHistory(messages []Message) *Builder` - restore conversation

### 7.2 Context Window Management ✅
- [x] `WithMaxHistory(n int) *Builder` - limit history
- [x] Auto-truncate old messages in addMessage()
- [x] System message preserved (not counted in history limit)

### 7.3 Tests ✅
- [x] Test GetHistory (returns copy, not reference)
- [x] Test SetHistory (replace history)
- [x] Test Clear (preserves system prompt)
- [x] Test WithMaxHistory (setting limit)
- [x] Test auto-truncate with maxHistory (FIFO removal)
- [x] Test unlimited history (maxHistory = 0)
- [x] Test method chaining (7 tests total)

### 7.4 Examples ✅
- [x] examples/builder_conversation.go (6 comprehensive examples)
  - Basic memory usage
  - Get and set history
  - Clear conversation
  - Max history limit with auto-truncation
  - Save and restore session
  - Memory vs no memory comparison

### 7.5 Documentation ✅
- [x] Updated BUILDER_API.md with Conversation Management section
- [x] Added to Table of Contents

---

## Phase 8: Error Handling & Recovery ✅ COMPLETE

### 8.1 Retry Logic ✅
- [x] `WithRetry(maxRetries int) *Builder`
- [x] `WithRetryDelay(delay time.Duration) *Builder`
- [x] `WithExponentialBackoff() *Builder`
- [x] Exponential backoff implementation (1s, 2s, 4s, 8s, 16s...)
- [x] Smart retryable error detection (rate limit, timeout)
- [x] `executeWithRetry` helper method

### 8.2 Timeout Management ✅
- [x] `WithTimeout(duration time.Duration) *Builder`
- [x] Wrap context with timeout automatically
- [x] Timeout detection and proper error wrapping
- [x] Context cancellation support

### 8.3 Error Types ✅
- [x] Define custom error types in agent/errors.go
  - [x] `ErrAPIKey` - missing/invalid key
  - [x] `ErrRateLimit` - rate limit exceeded
  - [x] `ErrTimeout` - request timeout
  - [x] `ErrRefusal` - content refused
  - [x] `ErrInvalidResponse` - malformed response
  - [x] `ErrMaxRetries` - max retry attempts exceeded
  - [x] `ErrToolExecution` - tool execution failed
- [x] `APIError` struct with type, message, status code
- [x] Error checking functions (IsAPIKeyError, IsRateLimitError, etc.)
- [x] Error wrapping functions (WrapAPIKey, WrapRateLimit, etc.)

### 8.4 Tests ✅
- [x] Test all custom error types (7 types)
- [x] Test APIError struct and Error() method
- [x] Test error checker functions (IsXXXError) - 27 test cases
- [x] Test WithTimeout, WithRetry, WithRetryDelay configuration
- [x] Test WithExponentialBackoff
- [x] Test calculateRetryDelay (fixed and exponential)
- [x] Test isRetryable error detection
- [x] Test executeWithRetry (14 tests total):
  - Success without retries
  - Eventual success with retries
  - Max retries exceeded
  - Non-retryable errors
  - Timeout handling
  - Method chaining

### 8.5 Examples ✅
- [x] examples/builder_errors.go (6 comprehensive examples)
  - Basic timeout handling
  - Retry with fixed delay
  - Retry with exponential backoff
  - Error type checking
  - Timeout with retry combination
  - Production-ready configuration
  - Bonus: Custom error handling wrapper

### 8.6 Documentation ✅
- [x] Updated BUILDER_API.md with Error Handling & Recovery section
- [x] Added to Table of Contents
- [x] Examples for timeout, retry, exponential backoff
- [x] Error type checking guide
- [x] Production-ready configuration example

---

## Phase 9: Examples & Documentation

### 9.1 Update Examples
- [ ] Rewrite `examples/ollama_example.go` with Builder API
- [ ] Rewrite `main.go` with Builder API
- [ ] Create `examples/builder_features.go` - showcase all features
- [ ] Create `examples/streaming_advanced.go` - streaming demo
- [ ] Create `examples/tools_demo.go` - tool calling demo
- [ ] Create `examples/json_schema_demo.go` - structured outputs

### 9.2 Update Documentation
- [ ] Update `README.md`
  - [ ] New quick start with Builder
  - [ ] All examples updated
  - [ ] Feature showcase
- [ ] Update `agent/README.md`
  - [ ] Full Builder API reference
  - [ ] All methods documented
- [ ] Update `ARCHITECTURE.md`
  - [ ] Add Builder pattern section
  - [ ] Explain design decisions
- [ ] Update `QUICK_REFERENCE.md`
  - [ ] New API examples
  - [ ] Common patterns

### 9.3 Migration Guide
- [ ] Create `MIGRATION.md`
  - [ ] Old API → New API mapping
  - [ ] Code examples before/after
  - [ ] Breaking changes list
  - [ ] Timeline (if phased approach)

---

## Phase 10: Testing & Quality

### 10.1 Unit Tests
- [ ] Test coverage > 80%
- [ ] All public methods tested
- [ ] Error cases covered
- [ ] Mock openai client for offline tests

### 10.2 Integration Tests
- [ ] Test with real OpenAI API (requires key)
- [ ] Test with real Ollama instance
- [ ] Test all features end-to-end

### 10.3 Benchmarks
- [ ] Benchmark simple requests
- [ ] Benchmark streaming
- [ ] Benchmark conversation management
- [ ] Compare with old API performance

### 10.4 Code Quality
- [ ] Run `go vet`
- [ ] Run `golangci-lint`
- [ ] Check code coverage
- [ ] Review all TODOs in code

---

## Phase 11: Advanced Features (Future)

### 11.1 RAG Support (Retrieval-Augmented Generation)
- [ ] `Retriever` interface
- [ ] `WithRAG(retriever) *Builder`
- [ ] Auto-retrieve context before query
- [ ] Embed retrieval results in prompt

### 11.2 Caching
- [ ] `WithCache(ttl time.Duration) *Builder`
- [ ] Cache responses by prompt hash
- [ ] Support cache invalidation
- [ ] Configurable cache backend (memory, redis)

### 11.3 Multimodal Support
- [ ] `WithImage(url string) *Builder`
- [ ] `WithAudio(url string) *Builder`
- [ ] Support vision models
- [ ] Support audio inputs

### 11.4 Batch Processing
- [ ] `AskBatch(ctx, messages []string) ([]string, error)`
- [ ] Parallel request processing
- [ ] Rate limit handling
- [ ] Progress callback

### 11.5 Chain of Thought
- [ ] `WithChainOfThought() *Builder`
- [ ] Auto-prompt for reasoning
- [ ] Extract reasoning steps
- [ ] Separate reasoning from answer

---

## Phase 12: Release Preparation

### 12.1 Version Management
- [ ] Update version to v2.0.0
- [ ] Tag release in git
- [ ] Update go.mod version

### 12.2 Documentation
- [ ] Complete API documentation
- [ ] Video tutorials (optional)
- [ ] Blog post announcement
- [ ] Update CHANGELOG.md

### 12.3 Community
- [ ] Announce on Reddit r/golang
- [ ] Post on Go Forum
- [ ] Tweet announcement
- [ ] Submit to awesome-go

---

## Metrics & Success Criteria

### Code Quality
- [ ] Test coverage > 80%
- [ ] Zero linter warnings
- [ ] All examples working
- [ ] Documentation complete

### Performance
- [ ] No performance regression vs old API
- [ ] Streaming latency < 100ms
- [ ] Memory usage reasonable

### User Experience
- [ ] Simple chat in 1 line of code
- [ ] No need to import openai-go
- [ ] Clear error messages
- [ ] Intuitive API

---

## Timeline Estimate

| Phase | Effort | Dependencies |
|-------|--------|--------------|
| Phase 1: Core Builder | 2-3 days | None |
| Phase 2: Advanced Params | 1 day | Phase 1 |
| Phase 3: Streaming | 2 days | Phase 1 |
| Phase 4: Tool Calling | 2 days | Phase 1, 3 |
| Phase 5: JSON Schema | 1 day | Phase 1 |
| Phase 6: Observability | 1 day | Phase 1 |
| Phase 7: Conversation | 1 day | Phase 1 |
| Phase 8: Error Handling | 1 day | Phase 1 |
| Phase 9: Examples & Docs | 2-3 days | All above |
| Phase 10: Testing | 2-3 days | All above |
| Phase 11: Advanced Features | 5-7 days | Phase 10 (optional) |
| Phase 12: Release | 1 day | Phase 10 |

**Total (Core + Polish): ~15-20 days**
**Total (With Advanced): ~20-27 days**

---

## Notes

- Start with Phase 1-8 for MVP (Minimum Viable Product)
- Phase 9-10 required before release
- Phase 11 can be done later as v2.1, v2.2, etc.
- Keep old API for reference but don't maintain (deprecated immediately)
- Focus on UX: every method should feel natural

---

## Current Status

**Started:** [Date when work begins]
**Current Phase:** Phase 0 - Planning
**Progress:** 0/12 phases complete

---

## Daily Progress Log

### [Date]
- [ ] Task completed
- Issues encountered:
- Decisions made:
- Next steps:

---

## Questions to Answer

1. Should we support Go 1.21+ or keep 1.23+?
2. Should Ask() panic or return error? (Currently: panic for simplicity)
3. Cache backend: memory-only or pluggable?
4. RAG retriever interface design?
5. How to handle tool calling loops (max iterations)?

---

## References

- [openai-go v3.8.1 docs](https://pkg.go.dev/github.com/openai/openai-go/v3)
- [OpenAI API reference](https://platform.openai.com/docs/api-reference)
- [Builder Pattern in Go](https://refactoring.guru/design-patterns/builder/go/example)
- [Fluent Interface](https://en.wikipedia.org/wiki/Fluent_interface)
