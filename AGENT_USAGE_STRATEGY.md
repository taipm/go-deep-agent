# Hướng Dẫn Sử Dụng Các Agent `bmad` Để Thực Thi Chiến Lược Phát Triển `go-deep-agent`

Tài liệu này mô tả quy trình và thứ tự sử dụng các agent chuyên biệt của `bmad` để thực hiện các công việc đã được vạch ra trong tài liệu `STRATEGY_NEXT_STEPS.md`.

---

## 🚀 Giai Đoạn 1: The Great Refactoring (Ưu tiên #1)

Đây là giai đoạn đòi hỏi sự kết hợp chặt chẽ giữa các vai trò kỹ thuật.

### Bước 1: Lập Kế Hoạch & Thiết Kế (Planning & Architecture)

1.  **`bmad-agent-bmm-pm` (Quản lý dự án)**
    -   **Mục đích:** Khởi động dự án và tạo kế hoạch chi tiết.
    -   **Câu lệnh mẫu:** *"Dựa vào 'Phase 1' trong `STRATEGY_NEXT_STEPS.md`, hãy tạo một danh sách các task cụ thể (task breakdown) cho việc refactor, bao gồm các đầu việc chính và ước tính thời gian cho mỗi việc."*
    -   **Kết quả:** Một backlog công việc rõ ràng (ví dụ: "Tạo interface LLMAdapter", "Implement OpenAI adapter", "Viết unit test cho Gemini adapter", v.v.).

2.  **`bmad-agent-bmm-architect` (Kiến trúc sư phần mềm)**
    -   **Mục đích:** Hoàn thiện thiết kế kỹ thuật trước khi viết code.
    -   **Câu lệnh mẫu:** *"Dựa trên "Thin Adapter Pattern" đã đề xuất, hãy thiết kế chi tiết `LLMAdapter` interface và các struct `CompletionRequest`, `CompletionResponse` trong Go. Chú ý đến các kiểu dữ liệu và comment giải thích."*
    -   **Kết quả:** Code interface và các struct dữ liệu sẵn sàng để implement.

### Bước 2: Phát Triển & Kiểm Thử (Development & Testing)

3.  **`bmad-agent-bmm-dev` (Lập trình viên)**
    -   **Mục đích:** Viết code cho các thành phần đã được thiết kế. Đây là agent bạn sẽ sử dụng nhiều nhất trong giai đoạn này.
    -   **Câu lệnh mẫu:**
        -   *"Implement `OpenAIAdapter` dựa trên interface `LLMAdapter` và wrap logic hiện có."*
        -   *"Bây giờ, implement `GeminiAdapter` sử dụng `generative-ai-go` SDK."*
        -   *"Refactor `Builder` để thay thế `*openai.Client` bằng `LLMAdapter`."*

4.  **`bmad-agent-bmm-tea` (Cố vấn kỹ thuật xuất sắc)**
    -   **Mục đích:** Đảm bảo chất lượng code và đưa ra các giải pháp tối ưu. Sử dụng xen kẽ với `bmm-dev`.
    -   **Câu lệnh mẫu:**
        -   *"Review code của `GeminiAdapter`. Có cách nào để xử lý việc chuyển đổi message format hiệu quả hơn không?"*
        -   *"Đề xuất một chiến lược viết unit test hiệu quả cho các adapter, bao gồm cả việc sử dụng mock."*

---

## ⚙️ Giai Đoạn 2: Polish & Production Ready

Giai đoạn này tập trung vào việc làm cho thư viện trở nên chuyên nghiệp và đáng tin cậy.

5.  **`bmad-agent-bmm-dev` (với vai trò DevOps)**
    -   **Mục đích:** Tự động hóa quy trình.
    -   **Câu lệnh mẫu:** *"Tạo một file GitHub Actions workflow để tự động chạy `go test ./...` và `golangci-lint run` mỗi khi có pull request vào nhánh `main`."*

6.  **`bmad-agent-bmm-tea` (Cố vấn kỹ thuật xuất sắc)**
    -   **Mục đích:** Nâng cao chất lượng và hiệu năng.
    -   **Câu lệnh mẫu:**
        -   *"Thiết kế một hệ thống error handling chuẩn hóa cho các adapter, định nghĩa các lỗi chung như `ErrRateLimit`, `ErrInvalidAPIKey`."*
        -   *"Viết code benchmark để so sánh performance của `Complete()` method trên 3 provider: OpenAI, Gemini, và Anthropic."*

---

## 🌱 Giai Đoạn 3: Growth & Community Building

Giai đoạn này tập trung vào việc truyền thông và xây dựng cộng đồng.

7.  **`bmad-agent-bmm-tech-writer` (Người viết tài liệu kỹ thuật)**
    -   **Mục đích:** Tạo ra tài liệu hấp dẫn và dễ hiểu.
    -   **Câu lệnh mẫu:**
        -   *"Cập nhật file `README.md`, thêm một section 'Multi-Provider Support' với các ví dụ code cho `NewOpenAI`, `NewGemini`, và `NewAnthropic`."*
        -   *"Viết một file `CONTRIBUTING.md` hướng dẫn cách để người khác có thể đóng góp vào dự án."*

8.  **`bmad-agent-cis-storyteller` (Người kể chuyện)**
    -   **Mục đích:** Tạo nội dung marketing hấp dẫn.
    -   **Câu lệnh mẫu:** *"Viết một bài blog với tiêu đề 'go-deep-agent 2.0: Hỗ trợ Gemini và Claude, Mở Ra Kỷ Nguyên Mới Cho AI trong Go' để công bố phiên bản mới."*

---

## 📋 Tóm Tắt Quy Trình Sử Dụng Agent

| Giai Đoạn | Công Việc Chính | Agent Chính | Agent Hỗ Trợ |
| :--- | :--- | :--- | :--- |
| **1. Refactoring** | Lập kế hoạch & Thiết kế | **`bmm-pm`**, **`bmm-architect`** | `bmm-tea` |
| | Phát triển & Test | **`bmm-dev`** | `bmm-tea` |
| **2. Polish** | CI/CD, Error Handling | **`bmm-dev`** | `bmm-tea` |
| **3. Growth** | Viết tài liệu, Marketing | **`bmm-tech-writer`** | `cis-storyteller` |

Bằng cách sử dụng các agent chuyên biệt theo đúng vai trò và thứ tự như trên, bạn sẽ mô phỏng được một quy trình làm việc chuyên nghiệp, giúp dự án phát triển một cách bài bản và hiệu quả nhất.
