# Memory Fix - Math Teacher Example

## Vấn đề phát hiện

Agent **không nhớ được cuộc hội thoại** trong interactive mode.

### Ví dụ lỗi:

```
👧 Con hỏi: Giải phương trình 3x^2 - 4x - 3 = 0
👩‍🏫 Cô giáo: [Bắt đầu giải...]

👧 Con hỏi: Vậy thì giải từng bước bài toán trên đi
👩‍🏫 Cô giáo: Hãy giải một phương trình đơn giản...  ← QUÊN phương trình 3x^2-4x-3=0
```

## Nguyên nhân

**Lỗi của thư viện:** `WithDefaults()` không bật `autoMemory`

- Documentation nói: "Memory(20): Keeps last 20 messages"
- Thực tế: Chỉ gọi `WithMaxHistory(20)`, không gọi `WithMemory()`
- Kết quả: Messages không được lưu vào history

Chi tiết: [BUG_REPORT_MEMORY_WITHDEFAULTS.md](../../BUG_REPORT_MEMORY_WITHDEFAULTS.md)

## Giải pháp

### Đã sửa trong example này:

**Trước (SAI):**
```go
teacher := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithDefaults().          // Không bật memory!
    WithPersona(persona).
    WithTools(...).
    WithMaxHistory(20)       // Vô dụng nếu không có memory
```

**Sau (ĐÚNG):**
```go
teacher := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithDefaults().
    WithMemory().            // ← QUAN TRỌNG: Bật memory
    WithPersona(persona).
    WithTools(...)
```

## Test lại

Sau khi fix, agent giờ đã nhớ được:

```
👧 Con hỏi: Tên con là Lan
👩‍🏫 Cô giáo: Chào Lan! Rất vui được gặp con.

👧 Con hỏi: Bạn nhớ tên con chưa?
👩‍🏫 Cô giáo: Dĩ nhiên rồi! Tên con là Lan.  ← NHỚ được!
```

## Khuyến nghị cho users khác

Nếu bạn dùng `WithDefaults()` và cần memory, **luôn thêm `.WithMemory()`**:

```go
// ❌ SAI - Memory không hoạt động
ai := agent.NewOpenAI(apiKey).WithDefaults()

// ✅ ĐÚNG - Memory hoạt động
ai := agent.NewOpenAI(apiKey).
    WithDefaults().
    WithMemory()
```

## Next steps

Bug đã được report cho tác giả thư viện. Sẽ được fix trong v0.7.10 hoặc v0.8.0.

---

**Fixed:** 2025-11-12
**Status:** ✅ Example đã hoạt động chính xác
