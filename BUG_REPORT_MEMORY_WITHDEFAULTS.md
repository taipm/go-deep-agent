# Bug Report: WithDefaults() không bật Memory

## Tóm tắt

`WithDefaults()` không bật `autoMemory`, dẫn đến agent **không nhớ được cuộc hội thoại** mặc dù documentation nói rằng "Memory(20): Keeps last 20 messages".

## Mức độ nghiêm trọng

🔴 **HIGH** - Ảnh hưởng đến trải nghiệm người dùng và vi phạm documentation

## Tái hiện lỗi

### Code

```go
package main

import (
    "context"
    "fmt"
    "github.com/taipm/go-deep-agent/agent"
)

func main() {
    ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
        WithDefaults()  // Documentation nói có Memory(20)

    // Hỏi tên
    ai.Ask(context.Background(), "Tên tôi là Alice")

    // Hỏi lại tên
    response, _ := ai.Ask(context.Background(), "Tên tôi là gì?")
    fmt.Println(response)
}
```

### Kết quả thực tế

```
Agent không nhớ tên Alice và trả lời: "Tôi không biết tên bạn là gì"
```

### Kết quả mong đợi

```
Agent nhớ và trả lời: "Tên bạn là Alice"
```

## Phân tích nguyên nhân

### 1. Documentation của `WithDefaults()` (builder_defaults.go:5-12)

```go
// WithDefaults configures the agent with production-ready default settings.
// This is the recommended starting point for most production use cases.
//
// Default Configuration:
//   - Memory(20): Keeps last 20 messages in conversation history    ← NÓI CÓ MEMORY
//   - Retry(3): Retries failed requests up to 3 times
//   - Timeout(30s): Sets 30-second timeout for API requests
//   - ExponentialBackoff: Uses exponential backoff for retries (1s, 2s, 4s, ...)
```

### 2. Implementation của `WithDefaults()` (builder_defaults.go:40-54)

```go
func (b *Builder) WithDefaults() *Builder {
    // Memory: Keep last 20 messages
    b.WithMaxHistory(20)    // ← CHỈ SET maxHistory, KHÔNG bật autoMemory

    // Retry: Retry failed requests up to 3 times
    b.WithRetry(3)

    // Timeout: 30-second timeout for API requests
    b.WithTimeout(30 * time.Second)

    // ExponentialBackoff: Smart retry delays (1s, 2s, 4s, ...)
    b.WithExponentialBackoff()

    return b
}
```

### 3. Cơ chế lưu messages (builder_execution.go:220-222)

```go
if b.autoMemory {
    b.addMessage(User(message))
    b.addMessage(Assistant(result))
}
```

Messages **CHỈ được lưu** khi `autoMemory == true`.

### 4. `WithMemory()` vs `WithMaxHistory()`

**WithMemory() - builder_memory.go:11-14**
```go
func (b *Builder) WithMemory() *Builder {
    b.autoMemory = true    // ← Bật memory
    return b
}
```

**WithMaxHistory() - builder_messages.go:77-80**
```go
func (b *Builder) WithMaxHistory(max int) *Builder {
    b.maxHistory = max    // ← CHỈ giới hạn số lượng, KHÔNG bật memory
    return b
}
```

## Vấn đề

1. ❌ `WithDefaults()` gọi `WithMaxHistory(20)` nhưng **KHÔNG gọi `WithMemory()`**
2. ❌ `autoMemory` vẫn là `false` (default value)
3. ❌ Messages **KHÔNG được lưu** vào history
4. ❌ Agent **KHÔNG NHỚ** được cuộc hội thoại
5. ❌ **Vi phạm documentation** - Doc nói "Memory(20)" nhưng thực tế không có memory

## Ảnh hưởng

### Use cases bị ảnh hưởng

1. **Chatbot**: Không nhớ tên người dùng, context cuộc trò chuyện
2. **Math Teacher**: Không nhớ các bài toán đã giải trước đó
3. **Customer Support**: Mất context giữa các câu hỏi
4. **Code Assistant**: Không nhớ code đã discuss

### Ví dụ thực tế (Math Teacher Example)

```
👧 Con hỏi: Giải phương trình 3x^2 - 4x - 3 = 0
👩‍🏫 Cô giáo: [Bắt đầu giải thích...]

👧 Con hỏi: Cô tính luôn cho con đi
👩‍🏫 Cô giáo: [Giải thích tiếp...]

👧 Con hỏi: Vậy thì giải từng bước bài toán giải phương trình trên đi
👩‍🏫 Cô giáo: Chắc chắn rồi! Hãy giải một phương trình đơn giản...
                ← QUÊN MẤT phương trình 3x^2 - 4x - 3 = 0

👧 Con hỏi: Ý con là phương trình bậc 2 con đã gửi cho cô
👩‍🏫 Cô giáo: Chào con! Rất vui khi nghe con nói về phương trình bậc 2...
                ← VẪN KHÔNG NHỚ phương trình cụ thể
```

## Giải pháp đề xuất

### Option 1: Sửa `WithDefaults()` để bật memory (KHUYẾN NGHỊ)

```go
func (b *Builder) WithDefaults() *Builder {
    // Memory: Keep last 20 messages
    b.WithMemory()           // ← THÊM dòng này
    b.WithMaxHistory(20)

    // Retry: Retry failed requests up to 3 times
    b.WithRetry(3)

    // Timeout: 30-second timeout for API requests
    b.WithTimeout(30 * time.Second)

    // ExponentialBackoff: Smart retry delays (1s, 2s, 4s, ...)
    b.WithExponentialBackoff()

    return b
}
```

**Ưu điểm:**
- ✅ Khớp với documentation
- ✅ UX tốt - memory enabled by default
- ✅ Sửa ở 1 chỗ, benefit cho tất cả users

**Nhược điểm:**
- ⚠️ Breaking change nếu ai đó rely vào behavior hiện tại

### Option 2: Cập nhật documentation (TẠM THỜI)

Sửa documentation để chính xác với implementation:

```go
// Default Configuration:
//   - MaxHistory(20): Limits conversation history to last 20 messages
//   - Retry(3): Retries failed requests up to 3 times
//   - Timeout(30s): Sets 30-second timeout for API requests
//   - ExponentialBackoff: Uses exponential backoff for retries (1s, 2s, 4s, ...)
//
// Note: Memory is NOT enabled by default. Call WithMemory() to enable:
//   agent.NewOpenAI(apiKey).WithDefaults().WithMemory()
```

**Ưu điểm:**
- ✅ Không breaking change

**Nhược điểm:**
- ❌ UX kém - users phải manually add `.WithMemory()`
- ❌ Confusing - "MaxHistory" nhưng không có history

### Option 3: Tách thành 2 methods

```go
// WithDefaults - production defaults WITHOUT memory
func (b *Builder) WithDefaults() *Builder {
    b.WithMaxHistory(20)
    b.WithRetry(3)
    b.WithTimeout(30 * time.Second)
    b.WithExponentialBackoff()
    return b
}

// WithDefaultsMemory - production defaults WITH memory (NEW)
func (b *Builder) WithDefaultsMemory() *Builder {
    return b.WithDefaults().WithMemory()
}
```

**Ưu điểm:**
- ✅ Backward compatible
- ✅ Clear naming

**Nhược điểm:**
- ❌ Thêm API surface
- ❌ Vẫn vi phạm doc của `WithDefaults()` hiện tại

## Khuyến nghị

🎯 **Option 1** - Sửa `WithDefaults()` để bật memory

**Lý do:**
1. Documentation đã commit rằng có "Memory(20)"
2. Đa số use cases cần memory (chatbot, assistant, tutor...)
3. Behavior hiện tại là unexpected và gây confusion
4. Breaking change có thể accept được vì:
   - Thư viện còn v0.x (chưa v1.0)
   - Fix một bug, không phải thay đổi behavior
   - Users chưa rely nhiều vào behavior sai này

**Migration path cho users:**
```go
// Nếu có ai đó MUỐN disable memory (rare case):
agent.NewOpenAI(apiKey).
    WithDefaults().
    DisableMemory()  // Opt-out
```

## Workaround hiện tại

Cho đến khi bug được fix, users cần manually add `.WithMemory()`:

```go
// ❌ SAI - Memory không hoạt động
ai := agent.NewOpenAI(apiKey).WithDefaults()

// ✅ ĐÚNG - Memory hoạt động
ai := agent.NewOpenAI(apiKey).
    WithDefaults().
    WithMemory()    // ← Phải thêm dòng này
```

## Test case

```go
func TestWithDefaultsEnablesMemory(t *testing.T) {
    ai := agent.NewOpenAI("gpt-4o-mini", apiKey).WithDefaults()

    // First message
    ai.Ask(context.Background(), "My name is Alice")

    // Second message - should remember
    response, _ := ai.Ask(context.Background(), "What is my name?")

    // Assert memory works
    assert.Contains(t, strings.ToLower(response), "alice")
}
```

## Files cần sửa

1. **agent/builder_defaults.go:40-54** - Thêm `b.WithMemory()`
2. **agent/builder_defaults_test.go** - Thêm test case cho memory
3. **CHANGELOG.md** - Document breaking change
4. **examples/** - Update examples nếu cần

## Timeline đề xuất

- **v0.7.10** - Fix bug này (breaking change nhưng justified)
- Release notes cần nói rõ:
  - What changed
  - Why it changed (bug fix)
  - Migration path

---

**Reporter:** AI Assistant
**Date:** 2025-11-12
**Version affected:** v0.7.9 và trước đó
**Priority:** HIGH
**Type:** Bug - Documentation mismatch + Unexpected behavior
