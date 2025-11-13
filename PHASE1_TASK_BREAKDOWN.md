# TASK BREAKDOWN - PHASE 1: THE GREAT REFACTORING

**Dự án:** go-deep-agent Multi-Provider Integration  
**Phase:** 1 - The Great Refactoring  
**Mục tiêu:** Tái cấu trúc để hỗ trợ OpenAI, Gemini, Anthropic thông qua "Thin Adapter Pattern"  
**Timeline:** 3-4 tuần (128 giờ)

---

## 📋 WEEK 1: FOUNDATION & OPENAI ADAPTER

### **Task 1.1: Thiết Kế Interface Core**
**Ước tính:** 4 giờ | **Priority:** P0 (Critical)

#### Subtasks:
- [ ] **1.1.1** Tạo file `agent/adapter.go` (1.5h)
  - Định nghĩa interface `LLMAdapter` với 2 methods: `Complete()` và `Stream()`
  - Viết documentation chi tiết cho interface
  
- [ ] **1.1.2** Định nghĩa struct `CompletionRequest` (1h)
  - Các trường: Model, Messages, System, Temperature, MaxTokens, Tools, TopP, Stop, Seed
  - Thêm comment giải thích từng trường
  
- [ ] **1.1.3** Định nghĩa struct `CompletionResponse` (1h)
  - Các trường: Content, ToolCalls, Usage, FinishReason
  - Định nghĩa struct `TokenUsage`
  
- [ ] **1.1.4** Review và test compile (0.5h)
  - Chạy `go build ./agent/...` để kiểm tra syntax

**Deliverable:** File `agent/adapter.go` với interface đầy đủ

---

### **Task 1.2: Tạo Package Adapters**
**Ước tính:** 2 giờ | **Priority:** P0 (Critical)

#### Subtasks:
- [ ] **1.2.1** Tạo cấu trúc thư mục (0.5h)
  - Tạo `agent/adapters/` directory
  - Tạo file stub: `openai_adapter.go`, `gemini_adapter.go`, `anthropic_adapter.go`
  
- [ ] **1.2.2** Setup package và imports (0.5h)
  - Khai báo package `adapters`
  - Import các dependency cần thiết
  
- [ ] **1.2.3** Tạo test files (1h)
  - `openai_adapter_test.go`, `gemini_adapter_test.go`, `anthropic_adapter_test.go`
  - Setup test structure cơ bản

**Deliverable:** Cấu trúc thư mục `agent/adapters/` hoàn chỉnh

---

### **Task 1.3: Implement OpenAI Adapter**
**Ước tính:** 12 giờ | **Priority:** P0 (Critical)

#### Subtasks:
- [ ] **1.3.1** Tạo struct `OpenAIAdapter` (1h)
  - Wrap `*openai.Client`
  - Constructor `NewOpenAIAdapter(apiKey, baseURL)`
  
- [ ] **1.3.2** Implement method `Complete()` (4h)
  - Convert `CompletionRequest` → OpenAI params
  - Handle system prompt (as message)
  - Convert messages array
  - Add tools nếu có
  - Set temperature, maxTokens, topP, stop, seed
  - Call OpenAI API
  
- [ ] **1.3.3** Implement conversion helpers (3h)
  - `convertMessages()`: agent.Message → openai.ChatCompletionMessageParamUnion
  - `convertTools()`: agent.Tool → openai.ChatCompletionToolParam
  - `convertResponse()`: openai.ChatCompletion → agent.CompletionResponse
  - `convertToolCalls()`: handle tool call conversion
  
- [ ] **1.3.4** Implement method `Stream()` (3h)
  - Build params (tương tự Complete)
  - Sử dụng `client.Chat.Completions.NewStreaming()`
  - Process stream với `ChatCompletionAccumulator`
  - Call `onChunk` callback cho mỗi content chunk
  - Return final response
  
- [ ] **1.3.5** Error handling và edge cases (1h)
  - Handle nil values
  - Handle empty responses
  - Proper error wrapping

**Deliverable:** `openai_adapter.go` hoàn chỉnh (~200 lines)

---

### **Task 1.4: Unit Tests cho OpenAI Adapter**
**Ước tính:** 8 giờ | **Priority:** P1 (High)

#### Subtasks:
- [ ] **1.4.1** Test `Complete()` method (3h)
  - Test basic completion
  - Test với system prompt
  - Test với tools
  - Test với various parameters (temperature, maxTokens, etc.)
  
- [ ] **1.4.2** Test `Stream()` method (2h)
  - Test streaming response
  - Test với callback
  - Verify accumulated response
  
- [ ] **1.4.3** Test conversion helpers (2h)
  - Test message conversion
  - Test tool conversion
  - Test response conversion
  
- [ ] **1.4.4** Test error cases (1h)
  - Test với invalid params
  - Test với API errors
  - Test với nil values

**Deliverable:** `openai_adapter_test.go` với coverage >80%

---

### **Task 1.5: Refactor Builder Struct**
**Ước tính:** 4 giờ | **Priority:** P0 (Critical)

#### Subtasks:
- [ ] **1.5.1** Update `Builder` struct trong `agent/builder.go` (0.5h)
  - Thay thế field `client *openai.Client` → `adapter LLMAdapter`
  - Keep tất cả fields khác unchanged
  
- [ ] **1.5.2** Update Provider constants trong `agent/config.go` (0.5h)
  - Thêm: `ProviderGemini`, `ProviderAnthropic`
  - Giữ nguyên: `ProviderOpenAI`, `ProviderOllama`
  
- [ ] **1.5.3** Update constructors (1.5h)
  - `NewOpenAI()` - keep interface unchanged
  - `NewGemini()` - new constructor
  - `NewAnthropic()` - new constructor
  - `NewWithAdapter()` - for custom adapters
  
- [ ] **1.5.4** Verify compilation (1.5h)
  - `go build ./agent`
  - Fix any compilation errors

**Deliverable:** `builder.go` với adapter field

---

### **Task 1.6: Implement `ensureAdapter()` Logic**
**Ước tính:** 6 giờ | **Priority:** P0 (Critical)

#### Subtasks:
- [ ] **1.6.1** Tạo function `ensureAdapter()` trong `builder_execution.go` (2h)
  - Replace `ensureClient()`
  - Logic khởi tạo adapter dựa vào provider
  
- [ ] **1.6.2** Handle từng provider (2h)
  - Case `ProviderOpenAI`: `NewOpenAIAdapter(apiKey, baseURL)`
  - Case `ProviderOllama`: `NewOpenAIAdapter(apiKey, baseURL)` với baseURL
  - Case `ProviderGemini`: `NewGeminiAdapter(apiKey)` (placeholder)
  - Case `ProviderAnthropic`: `NewAnthropicAdapter(apiKey)` (placeholder)
  - Case `ProviderAzure`: `NewOpenAIAdapter()` với Azure config
  
- [ ] **1.6.3** Error handling (1h)
  - Proper error messages cho từng provider
  - Handle adapter creation failures
  
- [ ] **1.6.4** Testing (1h)
  - Test initialization cho mỗi provider
  - Test error cases

**Deliverable:** `ensureAdapter()` function hoàn chỉnh

---

### **Task 1.7: Refactor `Ask()` Method**
**Ước tính:** 8 giờ | **Priority:** P0 (Critical)

#### Subtasks:
- [ ] **1.7.1** Tạo helper `buildCompletionRequest()` (2h)
  - Convert Builder state → `CompletionRequest`
  - Include: model, system, messages, temperature, maxTokens, tools, topP, stop, seed
  
- [ ] **1.7.2** Refactor `Ask()` logic (3h)
  - Call `ensureAdapter()`
  - Build `CompletionRequest` từ Builder state
  - Call `adapter.Complete(ctx, req)`
  - Process response (tool calls, auto-execute)
  - Update conversation history nếu `autoMemory` enabled
  
- [ ] **1.7.3** Maintain existing features (2h)
  - Cache checking (before adapter call)
  - Rate limiting
  - Memory system
  - RAG retrieval
  - Tool execution loop
  
- [ ] **1.7.4** Testing với mock adapter (1h)
  - Create `MockAdapter` for testing
  - Test `Ask()` logic với mock
  - Verify tool execution

**Deliverable:** `Ask()` method refactored, tất cả features hoạt động

---

### **Task 1.8: Refactor `Stream()` Method**
**Ước tính:** 6 giờ | **Priority:** P0 (Critical)

#### Subtasks:
- [ ] **1.8.1** Refactor `Stream()` logic (3h)
  - Call `ensureAdapter()`
  - Build `CompletionRequest`
  - Setup streaming callback wrapper
  - Call `adapter.Stream(ctx, req, onChunk)`
  - Process final response
  
- [ ] **1.8.2** Handle callbacks (2h)
  - Integrate với `onStream` callback
  - Handle `onToolCall` callback
  - Handle `onRefusal` callback
  
- [ ] **1.8.3** Testing (1h)
  - Test streaming với mock adapter
  - Verify callback execution
  - Test error handling

**Deliverable:** `Stream()` method refactored

---

### **Task 1.9: Regression Testing**
**Ước tính:** 12 giờ | **Priority:** P0 (Critical)

#### Subtasks:
- [ ] **1.9.1** Run existing test suite (2h)
  - `go test ./agent/... -v`
  - Identify failing tests
  
- [ ] **1.9.2** Fix failing tests (6h)
  - Update tests để work với adapter
  - Fix any broken functionality
  
- [ ] **1.9.3** Verify all features still work (3h)
  - Test tool calling
  - Test streaming
  - Test memory system
  - Test RAG
  - Test caching
  - Test rate limiting
  - Test batch processing
  
- [ ] **1.9.4** Performance testing (1h)
  - Benchmark adapter overhead
  - Ensure no performance regression

**Deliverable:** Tất cả existing tests pass, zero breaking changes

---

## 📋 WEEK 2: GEMINI & ANTHROPIC ADAPTERS

### **Task 2.1: Add Gemini SDK Dependency**
**Ước tính:** 1 giờ | **Priority:** P0

#### Subtasks:
- [ ] **2.1.1** Update `go.mod` (0.5h)
  - Add `github.com/google/generative-ai-go`
  - Add `google.golang.org/api`
  - Run `go mod tidy`
  
- [ ] **2.1.2** Verify dependency (0.5h)
  - Test import
  - Check version compatibility

**Deliverable:** Gemini SDK dependency added

---

### **Task 2.2: Implement Gemini Adapter**
**Ước tính:** 14 giờ | **Priority:** P0

#### Subtasks:
- [ ] **2.2.1** Tạo `GeminiAdapter` struct (1h)
  - Wrap `*genai.Client`
  - Constructor với error handling
  
- [ ] **2.2.2** Implement `configureModel()` helper (3h)
  - Set `SystemInstruction` (Gemini-specific)
  - Handle temperature (clamp to 0-1)
  - Set maxTokens, topP, stop sequences
  - Convert tools
  
- [ ] **2.2.3** Implement message conversion (2h)
  - Gemini uses different format (parts)
  - Handle role mapping (user → user, assistant → model)
  - Convert content to `genai.Text` parts
  
- [ ] **2.2.4** Implement `Complete()` method (3h)
  - Create model with `client.GenerativeModel()`
  - Configure model
  - Convert messages
  - Call `GenerateContent()`
  - Convert response
  
- [ ] **2.2.5** Implement `Stream()` method (3h)
  - Use `GenerateContentStream()`
  - Process iterator
  - Extract text from parts
  - Call onChunk callback
  - Track usage metadata
  
- [ ] **2.2.6** Handle Gemini quirks (2h)
  - System instruction via `SystemInstruction`
  - Role "model" vs "assistant"
  - Temperature range 0-1
  - Parts-based content

**Deliverable:** `gemini_adapter.go` hoàn chỉnh (~200 lines)

---

### **Task 2.3: Unit Tests cho Gemini Adapter**
**Ước tính:** 8 giờ | **Priority:** P1

#### Subtasks:
- [ ] **2.3.1** Test basic completion (2h)
  - Test với simple prompt
  - Test với system instruction
  - Verify response format
  
- [ ] **2.3.2** Test streaming (2h)
  - Test iterator processing
  - Verify callback execution
  - Test usage metadata
  
- [ ] **2.3.3** Test message conversion (2h)
  - Test role mapping
  - Test parts conversion
  
- [ ] **2.3.4** Test edge cases (2h)
  - Test temperature clamping
  - Test empty responses
  - Test error handling

**Deliverable:** `gemini_adapter_test.go` với coverage >80%

---

### **Task 2.4: Integration Tests cho Gemini**
**Ước tính:** 4 giờ | **Priority:** P2

#### Subtasks:
- [ ] **2.4.1** Setup integration test (1h)
  - Require API key via env var
  - Skip if key not present
  
- [ ] **2.4.2** Test real API calls (2h)
  - Test basic completion
  - Test streaming
  - Test với tools (if supported)
  
- [ ] **2.4.3** Verify với Builder (1h)
  - Test `NewGemini()` constructor
  - Test end-to-end flow

**Deliverable:** Working Gemini integration

---

### **Task 2.5: Add Anthropic SDK Dependency**
**Ước tính:** 1 giờ | **Priority:** P0

#### Subtasks:
- [ ] **2.5.1** Update `go.mod` (0.5h)
  - Add `github.com/anthropics/anthropic-sdk-go`
  - Run `go mod tidy`
  
- [ ] **2.5.2** Verify dependency (0.5h)
  - Test import
  - Check compatibility

**Deliverable:** Anthropic SDK dependency added

---

### **Task 2.6: Implement Anthropic Adapter**
**Ước tính:** 14 giờ | **Priority:** P0

#### Subtasks:
- [ ] **2.6.1** Tạo `AnthropicAdapter` struct (1h)
  - Wrap `*anthropic.Client`
  - Constructor
  
- [ ] **2.6.2** Implement message conversion (3h)
  - NO system role in messages!
  - Convert user/assistant roles
  - Handle tool results
  
- [ ] **2.6.3** Implement `buildParams()` helper (3h)
  - Set system via separate parameter
  - Set maxTokens (REQUIRED!)
  - Handle temperature (clamp to 0-1)
  - Set topP, stop sequences
  - Convert tools
  
- [ ] **2.6.4** Implement `Complete()` method (3h)
  - Build params
  - Call `client.Messages.New()`
  - Convert response
  - Extract content from blocks
  - Extract usage
  
- [ ] **2.6.5** Implement `Stream()` method (3h)
  - Use `NewStreaming()`
  - Process event stream
  - Handle `content_block_delta` events
  - Handle `message_delta` events
  - Call onChunk callback
  
- [ ] **2.6.6** Handle Anthropic quirks (1h)
  - `maxTokens` REQUIRED
  - System as separate param
  - Content blocks format
  - Different parameter names

**Deliverable:** `anthropic_adapter.go` hoàn chỉnh (~200 lines)

---

### **Task 2.7: Unit Tests cho Anthropic Adapter**
**Ước tính:** 8 giờ | **Priority:** P1

#### Subtasks:
- [ ] **2.7.1** Test basic completion (2h)
  - Test message conversion
  - Test system handling
  - Verify maxTokens requirement
  
- [ ] **2.7.2** Test streaming (2h)
  - Test event processing
  - Test content_block_delta
  - Verify callback
  
- [ ] **2.7.3** Test response conversion (2h)
  - Test content block extraction
  - Test usage extraction
  
- [ ] **2.7.4** Test edge cases (2h)
  - Test missing maxTokens
  - Test temperature clamping
  - Test error handling

**Deliverable:** `anthropic_adapter_test.go` với coverage >80%

---

### **Task 2.8: Integration Tests cho Anthropic**
**Ước tính:** 4 giờ | **Priority:** P2

#### Subtasks:
- [ ] **2.8.1** Setup integration test (1h)
  - Require API key
  - Skip if not present
  
- [ ] **2.8.2** Test real API calls (2h)
  - Test basic completion
  - Test streaming
  - Test với tools
  
- [ ] **2.8.3** Verify với Builder (1h)
  - Test `NewAnthropic()` constructor
  - Test end-to-end

**Deliverable:** Working Anthropic integration

---

### **Task 2.9: Cross-Provider Testing**
**Ước tính:** 6 giờ | **Priority:** P1

#### Subtasks:
- [ ] **2.9.1** Tạo cross-provider test suite (2h)
  - Test cùng prompt trên 3 providers
  - Verify tất cả đều trả về response
  
- [ ] **2.9.2** Test feature parity (3h)
  - Test tool calling trên cả 3
  - Test streaming trên cả 3
  - Test parameters (temperature, maxTokens)
  
- [ ] **2.9.3** Document differences (1h)
  - Note provider-specific behaviors
  - Document quirks

**Deliverable:** Cross-provider test suite

---

### **Task 2.10: Performance & Polish**
**Ước tính:** 6 giờ | **Priority:** P2

#### Subtasks:
- [ ] **2.10.1** Benchmarking (2h)
  - Benchmark adapter overhead
  - Compare performance across providers
  
- [ ] **2.10.2** Error handling improvements (2h)
  - Standardize error messages
  - Add proper error wrapping
  
- [ ] **2.10.3** Code cleanup (2h)
  - Remove commented code
  - Improve naming
  - Add missing comments

**Deliverable:** Polished, production-ready code

---

## 📊 SUMMARY & METRICS

### Total Estimated Time
- **Week 1 (Foundation + OpenAI):** 62 giờ (≈8 ngày làm việc)
- **Week 2 (Gemini + Anthropic):** 66 giờ (≈8.5 ngày làm việc)
- **Total:** 128 giờ (≈16 ngày làm việc)

### Priority Breakdown
- **P0 (Critical):** 110 giờ (86%)
- **P1 (High):** 12 giờ (9%)
- **P2 (Medium):** 6 giờ (5%)

### Task Distribution
- **Implementation:** 72 giờ (56%)
- **Testing:** 46 giờ (36%)
- **Setup/Infra:** 10 giờ (8%)

### Critical Path
1. Task 1.1 → 1.2 → 1.3 → 1.5 → 1.6 → 1.7 → 1.9 (OpenAI adapter + Builder refactor)
2. Task 2.1 → 2.2 → 2.3 (Gemini)
3. Task 2.5 → 2.6 → 2.7 (Anthropic)
4. Task 2.9 → 2.10 (Cross-provider testing + Polish)

### Milestones
- **Milestone 1 (End of Week 1):** OpenAI adapter working, all existing tests passing
- **Milestone 2 (Mid Week 2):** Gemini adapter complete and tested
- **Milestone 3 (End of Week 2):** Anthropic adapter complete and tested
- **Milestone 4 (Final):** Cross-provider tests passing, documentation complete

### Risk Factors & Mitigation

| Risk | Impact | Probability | Mitigation | Buffer |
|------|--------|-------------|------------|--------|
| Gemini SDK quirks phức tạp | High | Medium | Research SDK trước, có fallback plan | +20% |
| Anthropic API khác biệt nhiều | High | Medium | Study documentation kỹ, test sớm | +20% |
| Existing tests cần refactor nhiều | Medium | High | Incremental refactoring, mock adapters | +15% |
| Dependency conflicts | Low | Low | Test dependency early | -5% |
| Performance regression | Medium | Low | Early benchmarking, profiling | +10% |

### Recommended Timeline
- **Realistic (1 developer):** 3-4 tuần full-time
- **Aggressive (2 developers):** 2-3 tuần với parallel work
- **Conservative (thorough testing):** 4-5 tuần với comprehensive QA

### Success Criteria
- [ ] All 3 adapters (OpenAI, Gemini, Anthropic) implemented
- [ ] All existing tests passing (zero breaking changes)
- [ ] Unit test coverage >80% for adapters
- [ ] Integration tests passing for all providers
- [ ] Cross-provider feature parity verified
- [ ] Performance overhead <1ms per request
- [ ] Documentation complete
- [ ] Code review passed

### Next Steps After Phase 1
1. Update README with multi-provider examples
2. Create provider-specific guides
3. Add examples for each provider
4. Prepare release notes
5. Plan Phase 2 (Polish & Production Ready)

---

**Document Version:** 1.0  
**Created:** November 13, 2025  
**Status:** Ready for Execution  
**Approver:** Project Lead
