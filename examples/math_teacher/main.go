package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/taipm/go-deep-agent/agent"
	"github.com/taipm/go-deep-agent/agent/tools"
)

// CreateMathTeacher tạo một giáo viên toán học AI với persona được cấu hình sẵn
func CreateMathTeacher(apiKey string) (*agent.Builder, error) {
	// Load persona từ file YAML
	// Tự động tìm file trong thư mục hiện tại hoặc từ root
	personaPath := "math_teacher.yaml"
	if _, err := os.Stat(personaPath); os.IsNotExist(err) {
		personaPath = "examples/math_teacher/math_teacher.yaml"
	}

	persona, err := agent.LoadPersona(personaPath)
	if err != nil {
		return nil, fmt.Errorf("không thể load persona: %w", err)
	}

	// Tạo agent với persona và các tools hữu ích
	teacher := agent.NewOpenAI("gpt-4o-mini", apiKey).
		WithDefaults().          // Memory(20) + Retry(3) + Timeout(30s) + ExponentialBackoff
		WithPersona(persona).    // Load tính cách giáo viên từ YAML
		WithTools(
			tools.NewMathTool(),     // Công cụ tính toán
			tools.NewDateTimeTool(), // Công cụ xử lý thời gian (cho bài toán có ngày tháng)
		).
		WithAutoExecute(true).   // Tự động thực thi các tools
		WithToolChoice("auto").  // Để AI quyết định khi nào dùng tools
		WithInfoLogging()        // Bật logging cho production

	return teacher, nil
}

// Example1: Câu hỏi đơn giản về phép cộng
func Example1_SimpleMath(teacher *agent.Builder) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("VÍ DỤ 1: BÀI TOÁN CỘNG ĐỐN GIẢN")
	fmt.Println(strings.Repeat("=", 60))

	question := "Con gái, hôm nay cô dạy con bài này nhé: 15 + 27 bằng bao nhiêu?"

	fmt.Printf("\n👧 Con hỏi: %s\n\n", question)

	response, err := teacher.Ask(context.Background(), question)
	if err != nil {
		fmt.Printf("❌ Lỗi: %v\n", err)
		return
	}

	fmt.Printf("👩‍🏫 Cô giáo: %s\n", response)
}

// Example2: Bài toán có lời văn
func Example2_WordProblem(teacher *agent.Builder) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("VÍ DỤ 2: BÀI TOÁN CÓ LỜI VĂN")
	fmt.Println(strings.Repeat("=", 60))

	question := "Mẹ mua cho con 3 hộp kẹo. Mỗi hộp có 5 viên kẹo. Hỏi con có tất cả bao nhiêu viên kẹo?"

	fmt.Printf("\n👧 Con hỏi: %s\n\n", question)

	response, err := teacher.Ask(context.Background(), question)
	if err != nil {
		fmt.Printf("❌ Lỗi: %v\n", err)
		return
	}

	fmt.Printf("👩‍🏫 Cô giáo: %s\n", response)
}

// Example3: Phân số
func Example3_Fractions(teacher *agent.Builder) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("VÍ DỤ 3: BÀI TOÁN PHÂN SỐ")
	fmt.Println(strings.Repeat("=", 60))

	question := "Cô ơi, 1/2 của 8 là bao nhiêu ạ? Con không hiểu lắm."

	fmt.Printf("\n👧 Con hỏi: %s\n\n", question)

	response, err := teacher.Ask(context.Background(), question)
	if err != nil {
		fmt.Printf("❌ Lỗi: %v\n", err)
		return
	}

	fmt.Printf("👩‍🏫 Cô giáo: %s\n", response)
}

// Example4: Phép tính phức tạp hơn
func Example4_ComplexCalculation(teacher *agent.Builder) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("VÍ DỤ 4: BÀI TOÁN PHỨC TẠP HƠN")
	fmt.Println(strings.Repeat("=", 60))

	question := "Con có 100 nghìn đồng. Con mua 3 quyển vở, mỗi quyển 8 nghìn. Sau đó con mua 2 cái bút, mỗi cái 5 nghìn. Hỏi con còn lại bao nhiêu tiền?"

	fmt.Printf("\n👧 Con hỏi: %s\n\n", question)

	response, err := teacher.Ask(context.Background(), question)
	if err != nil {
		fmt.Printf("❌ Lỗi: %v\n", err)
		return
	}

	fmt.Printf("👩‍🏫 Cô giáo: %s\n", response)
}

// Example5: Hình học cơ bản
func Example5_Geometry(teacher *agent.Builder) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("VÍ DỤ 5: BÀI TOÁN HÌNH HỌC")
	fmt.Println(strings.Repeat("=", 60))

	question := "Cô ơi, làm sao tính chu vi hình chữ nhật có chiều dài 10cm và chiều rộng 6cm?"

	fmt.Printf("\n👧 Con hỏi: %s\n\n", question)

	response, err := teacher.Ask(context.Background(), question)
	if err != nil {
		fmt.Printf("❌ Lỗi: %v\n", err)
		return
	}

	fmt.Printf("👩‍🏫 Cô giáo: %s\n", response)
}

// Example6: Chế độ tương tác - chat liên tục
func Example6_InteractiveMode(teacher *agent.Builder) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("VÍ DỤ 6: CHẾ ĐỘ TƯƠNG TÁC (Chat liên tục)")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("\n💡 Bạn có thể chat với cô giáo toán. Gõ 'exit' để thoát.\n")

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("👧 Con hỏi: ")
		if !scanner.Scan() {
			break
		}

		question := strings.TrimSpace(scanner.Text())

		if question == "" {
			continue
		}

		if strings.ToLower(question) == "exit" {
			fmt.Println("\n👩‍🏫 Cô giáo: Tạm biệt con! Học tốt nhé! ❤️")
			break
		}

		response, err := teacher.Ask(context.Background(), question)
		if err != nil {
			fmt.Printf("❌ Lỗi: %v\n\n", err)
			continue
		}

		fmt.Printf("\n👩‍🏫 Cô giáo: %s\n\n", response)
	}
}

func main() {
	// Lấy API key từ environment variable
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("❌ Lỗi: Vui lòng set OPENAI_API_KEY environment variable")
		fmt.Println("   Ví dụ: export OPENAI_API_KEY='sk-...'")
		os.Exit(1)
	}

	// Tạo giáo viên toán học
	fmt.Println("🎓 Đang khởi tạo Cô Giáo Toán học AI...")
	teacher, err := CreateMathTeacher(apiKey)
	if err != nil {
		fmt.Printf("❌ Lỗi khi tạo giáo viên: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Cô giáo đã sẵn sàng!\n")

	// Nếu có tham số dòng lệnh, chạy chế độ tương ứng
	if len(os.Args) > 1 {
		mode := os.Args[1]
		switch mode {
		case "1", "simple":
			Example1_SimpleMath(teacher)
		case "2", "word":
			Example2_WordProblem(teacher)
		case "3", "fraction":
			Example3_Fractions(teacher)
		case "4", "complex":
			Example4_ComplexCalculation(teacher)
		case "5", "geometry":
			Example5_Geometry(teacher)
		case "6", "interactive", "chat":
			Example6_InteractiveMode(teacher)
		default:
			fmt.Printf("❌ Không hiểu tham số: %s\n", mode)
			fmt.Println("📖 Sử dụng: go run main.go [1|2|3|4|5|6]")
			os.Exit(1)
		}
	} else {
		// Chạy tất cả các ví dụ
		fmt.Println("🚀 Chạy tất cả các ví dụ...\n")

		Example1_SimpleMath(teacher)
		Example2_WordProblem(teacher)
		Example3_Fractions(teacher)
		Example4_ComplexCalculation(teacher)
		Example5_Geometry(teacher)

		fmt.Println("\n" + strings.Repeat("=", 60))
		fmt.Println("🎉 HOÀN THÀNH TẤT CẢ CÁC VÍ DỤ!")
		fmt.Println(strings.Repeat("=", 60))
		fmt.Println("\n💡 Để thử chế độ chat tương tác, chạy:")
		fmt.Println("   go run main.go interactive")
		fmt.Println("\n💡 Hoặc chạy từng ví dụ riêng lẻ:")
		fmt.Println("   go run main.go 1  # Ví dụ 1: Phép cộng")
		fmt.Println("   go run main.go 2  # Ví dụ 2: Bài toán có lời văn")
		fmt.Println("   go run main.go 3  # Ví dụ 3: Phân số")
		fmt.Println("   go run main.go 4  # Ví dụ 4: Bài toán phức tạp")
		fmt.Println("   go run main.go 5  # Ví dụ 5: Hình học")
		fmt.Println("   go run main.go 6  # Ví dụ 6: Chat tương tác")
	}
}
