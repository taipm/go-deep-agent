# Quick Start - Giáo Viên Toán Học AI 🚀

Hướng dẫn nhanh để chạy example trong 3 bước.

## Bước 1: Cài đặt API Key

```bash
export OPENAI_API_KEY='sk-your-api-key-here'
```

> Lấy API key tại: https://platform.openai.com/api-keys

## Bước 2: Di chuyển vào thư mục

```bash
cd examples/math_teacher
```

## Bước 3: Chạy chương trình

### Chạy tất cả các ví dụ (khuyến nghị cho lần đầu)

```bash
go run .
```

Output sẽ hiển thị 5 ví dụ:
- ✅ Phép cộng đơn giản (15 + 27)
- ✅ Bài toán có lời văn (3 hộp kẹo × 5 viên)
- ✅ Phân số (1/2 của 8)
- ✅ Bài toán phức tạp (tính tiền còn lại)
- ✅ Hình học (chu vi hình chữ nhật)

### Chạy chế độ chat tương tác (thú vị nhất!)

```bash
go run . interactive
```

Sau đó bạn có thể chat tự do với cô giáo:

```
👧 Con hỏi: 15 + 27 bằng bao nhiêu?
👩‍🏫 Cô giáo: [Giải thích chi tiết từng bước]

👧 Con hỏi: Tại sao 5 × 4 = 4 × 5?
👩‍🏫 Cô giáo: [Giải thích tính chất giao hoán]

👧 Con hỏi: exit
👩‍🏫 Cô giáo: Tạm biệt con! Học tốt nhé! ❤️
```

### Chạy từng ví dụ riêng lẻ

```bash
go run . 1    # Phép cộng đơn giản
go run . 2    # Bài toán có lời văn
go run . 3    # Phân số
go run . 4    # Bài toán phức tạp
go run . 5    # Hình học
go run . 6    # Chat tương tác (giống "interactive")
```

## Các câu hỏi hay để thử

```
"15 + 27 bằng bao nhiêu?"
"Con có 3 hộp kẹo, mỗi hộp 5 viên. Tổng cộng bao nhiêu viên?"
"1/2 của 8 là bao nhiêu?"
"Làm sao tính chu vi hình chữ nhật dài 10cm, rộng 6cm?"
"Tại sao 5 × 4 = 4 × 5?"
"Làm sao nhân nhanh với 9?"
"Con có 100 nghìn, mua 3 vở 8 nghìn/quyển. Còn bao nhiêu?"
```

## Xem thêm

- [README.md](README.md) - Hướng dẫn chi tiết
- [EXAMPLE_OUTPUT.md](EXAMPLE_OUTPUT.md) - Ví dụ output thực tế
- [math_teacher.yaml](math_teacher.yaml) - Cấu hình persona
- [main.go](main.go) - Source code

## Gặp vấn đề?

### Lỗi: "OPENAI_API_KEY not set"

```bash
export OPENAI_API_KEY='sk-...'
```

### Lỗi: "cannot find package"

```bash
cd /path/to/go-deep-agent
go mod download
cd examples/math_teacher
go run .
```

### Muốn dùng model khác?

Sửa trong [main.go:29](main.go#L29):

```go
// Thay vì gpt-4o-mini
teacher := agent.NewOpenAI("gpt-4o", apiKey)
```

---

**Chúc bạn và con có trải nghiệm học toán thú vị! 🎓✨**
