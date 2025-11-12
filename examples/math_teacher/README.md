# Giáo Viên Toán Học AI 👩‍🏫

Ví dụ này minh họa cách xây dựng một giáo viên toán học AI tận tâm, kiên nhẫn để giúp con bạn học toán một cách thú vị và hiệu quả.

## Tính năng nổi bật

- **Giải thích từng bước**: Chia nhỏ bài toán phức tạp thành các bước đơn giản
- **Ví dụ thực tế**: Sử dụng ngữ cảnh gần gũi với trẻ (kẹo, đồ chơi, trái cây...)
- **Kiên nhẫn và khuyến khích**: Không bao giờ làm trẻ cảm thấy xấu hổ
- **Công cụ tính toán tích hợp**: Sử dụng MathTool để tính toán chính xác
- **Nhớ ngữ cảnh**: Ghi nhớ 20 câu hỏi gần nhất trong cuộc trò chuyện
- **Tương tác linh hoạt**: Hỗ trợ cả chế độ ví dụ và chat tương tác

## Kiến thức toán học

Cô giáo AI có thể giúp con với:

- Số học cơ bản (cộng, trừ, nhân, chia)
- Phân số và số thập phân
- Hình học cơ bản
- Bài toán có lời văn
- Tư duy logic và giải quyết vấn đề

## Cài đặt

### Yêu cầu

- Go 1.23 trở lên
- OpenAI API key

### Thiết lập

1. Clone repository và di chuyển vào thư mục:

```bash
cd go-deep-agent/examples/math_teacher
```

2. Cài đặt dependencies:

```bash
go mod download
```

3. Thiết lập API key:

```bash
export OPENAI_API_KEY='sk-your-api-key-here'
```

## Cách sử dụng

### Chạy tất cả các ví dụ

```bash
go run main.go
```

Lệnh này sẽ chạy 5 ví dụ minh họa:
1. Phép cộng đơn giản
2. Bài toán có lời văn
3. Phân số
4. Bài toán phức tạp (nhiều bước)
5. Hình học cơ bản

### Chạy từng ví dụ riêng lẻ

```bash
# Ví dụ 1: Phép cộng đơn giản
go run main.go 1

# Ví dụ 2: Bài toán có lời văn
go run main.go 2

# Ví dụ 3: Phân số
go run main.go 3

# Ví dụ 4: Bài toán phức tạp
go run main.go 4

# Ví dụ 5: Hình học
go run main.go 5
```

### Chế độ chat tương tác

```bash
go run main.go interactive
# hoặc
go run main.go 6
# hoặc
go run main.go chat
```

Trong chế độ này, bạn có thể chat liên tục với cô giáo. Gõ `exit` để thoát.

**Ví dụ chat:**

```
👧 Con hỏi: 15 + 27 bằng bao nhiêu?

👩‍🏫 Cô giáo: Chào con! Cô sẽ giúp con giải bài toán này nhé! 😊

Để tính 15 + 27, chúng ta có thể chia nhỏ như sau:

Bước 1: Chia 27 thành 20 và 7
  15 + 27 = 15 + 20 + 7

Bước 2: Tính 15 + 20
  15 + 20 = 35

Bước 3: Cộng thêm 7
  35 + 7 = 42

Vậy 15 + 27 = 42! 🎉

Con có hiểu cách làm không? Hay con muốn cô giải thích thêm?

👧 Con hỏi: exit

👩‍🏫 Cô giáo: Tạm biệt con! Học tốt nhé! ❤️
```

## Cấu trúc code

### 1. File Persona (`math_teacher.yaml`)

File này định nghĩa tính cách, phong cách dạy học và hướng dẫn cho giáo viên AI:

```yaml
name: "MathTeacher"
role: "Giáo Viên Toán Học Tận Tâm"
personality:
  tone: "thân thiện, kiên nhẫn và khuyến khích"
  traits:
    - kiên nhẫn
    - rõ ràng
    - nhiệt tình
guidelines:
  - "Luôn chia nhỏ bài toán phức tạp thành các bước đơn giản"
  - "Sử dụng ví dụ thực tế mà con có thể liên hệ"
  - "Khen ngợi khi con làm đúng"
```

### 2. Hàm `CreateMathTeacher`

Tạo agent với cấu hình production-ready:

```go
func CreateMathTeacher(apiKey string) (*agent.Builder, error) {
    persona, _ := agent.LoadPersona("examples/math_teacher/math_teacher.yaml")

    return agent.NewOpenAI("gpt-4o-mini", apiKey).
        WithDefaults().              // Memory + Retry + Timeout
        WithPersona(persona).        // Load persona
        WithTools(
            tools.NewMathTool(),     // Công cụ tính toán
            tools.NewDateTimeTool(), // Xử lý thời gian
        ).
        WithAutoExecute(true).       // Tự động dùng tools
        WithMaxHistory(20),          // Nhớ 20 tin nhắn
    nil
}
```

### 3. Các ví dụ giảng dạy

Mỗi ví dụ minh họa một tình huống học tập khác nhau:

- **Example1**: Phép cộng đơn giản
- **Example2**: Bài toán có lời văn (nhân)
- **Example3**: Phân số
- **Example4**: Bài toán nhiều bước (tính tiền còn lại)
- **Example5**: Hình học (chu vi hình chữ nhật)
- **Example6**: Chat tương tác liên tục

## Tính năng nâng cao

### Tích hợp Tools

Giáo viên sử dụng 2 tools:

1. **MathTool**: Tính toán chính xác các phép toán
2. **DateTimeTool**: Xử lý bài toán liên quan đến thời gian

```go
.WithTools(
    tools.NewMathTool(),
    tools.NewDateTimeTool(),
)
```

### Memory (Ghi nhớ ngữ cảnh)

Agent nhớ 20 tin nhắn gần nhất, giúp duy trì ngữ cảnh trong cuộc trò chuyện:

```go
.WithMaxHistory(20)
```

### Auto-execute Tools

Agent tự động thực thi tools khi cần, không yêu cầu xác nhận:

```go
.WithAutoExecute(true)
```

### Production-ready với `WithDefaults()`

Tự động cấu hình:
- Memory: 20 tin nhắn
- Retry: 3 lần
- Timeout: 30 giây
- Exponential backoff cho retry

## Tùy chỉnh

### Thay đổi model

```go
agent.NewOpenAI("gpt-4o", apiKey)  // Dùng GPT-4o cho câu trả lời tốt hơn
```

### Điều chỉnh temperature

```go
.WithTemperature(0.7)  // 0-2, càng cao càng sáng tạo
```

### Thay đổi max tokens

```go
.WithMaxTokens(3000)  // Câu trả lời dài hơn
```

### Thêm tools khác

```go
.WithTools(
    tools.NewMathTool(),
    tools.NewDateTimeTool(),
    tools.NewFileSystemTool(),  // Thêm tool đọc/ghi file
)
```

### Chỉnh sửa persona

Edit file `math_teacher.yaml` để thay đổi:
- Phong cách dạy học
- Tone giọng nói
- Hướng dẫn cụ thể
- Các ví dụ minh họa

## Mẹo sử dụng hiệu quả

1. **Bắt đầu với câu hỏi đơn giản**: Để con làm quen với cô giáo AI
2. **Khuyến khích hỏi "tại sao"**: Cô giáo sẽ giải thích sâu hơn
3. **Dùng ví dụ thực tế**: Đặt câu hỏi liên quan đến đời sống hàng ngày
4. **Chat liên tục**: Dùng interactive mode để duy trì ngữ cảnh
5. **Khen ngợi**: Cô giáo sẽ động viên khi con làm đúng

## Ví dụ câu hỏi hay

```
1. "15 + 27 bằng bao nhiêu?"
2. "Nếu con có 3 hộp kẹo, mỗi hộp 5 viên thì có bao nhiêu viên?"
3. "1/2 của 8 là bao nhiêu?"
4. "Con có 100 nghìn. Mua 3 vở 8 nghìn/quyển. Còn bao nhiêu?"
5. "Chu vi hình chữ nhật dài 10cm, rộng 6cm?"
6. "Tại sao 5 × 4 = 4 × 5?"
7. "Làm sao nhân nhanh với 9?"
```

## Khắc phục sự cố

### Lỗi: "không thể load persona"

Đảm bảo bạn đang chạy từ thư mục `examples/math_teacher`:

```bash
cd examples/math_teacher
go run .
```

Code tự động tìm file `math_teacher.yaml` trong thư mục hiện tại.

### Lỗi: "OPENAI_API_KEY not set"

Thiết lập API key:

```bash
export OPENAI_API_KEY='sk-your-api-key'
```

### Agent không dùng tools

Kiểm tra `WithAutoExecute(true)` đã được bật.

## Tham khảo

- [go-deep-agent Documentation](https://github.com/taipm/go-deep-agent)
- [Persona Configuration Guide](../../docs/persona.md)
- [Tools Documentation](../../docs/tools.md)
- [Examples](../)

## Giấy phép

MIT License - xem file LICENSE trong repository chính.

## Đóng góp

Mọi đóng góp đều được chào đón! Vui lòng tạo issue hoặc pull request.

---

**Chúc con học toán vui vẻ! 🎓📐✨**
