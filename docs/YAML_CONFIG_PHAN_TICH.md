# Phân tích YAML Config: Traditional vs Persona-Based

**Ngày**: 10/11/2025  
**Mục đích**: Quyết định cách tiếp cận YAML config cho go-deep-agent v0.6.2+  
**Mục tiêu**: Chọn cách thiết kế tốt nhất cho developer experience

---

## Mục lục

1. [Tóm tắt](#tóm-tắt)
2. [Cách 1: Traditional Config](#cách-1-traditional-config)
3. [Cách 2: Persona-Based Config](#cách-2-persona-based-config)
4. [So sánh chi tiết](#so-sánh-chi-tiết)
5. [Ví dụ từ các framework khác](#ví-dụ-từ-các-framework-khác)
6. [Phân tích trải nghiệm developer](#phân-tích-trải-nghiệm-developer)
7. [Đề xuất cuối cùng](#đề-xuất-cuối-cùng)

---

## Tóm tắt

**Câu hỏi**: go-deep-agent nên dùng traditional config hay persona-based config?

**Trả lời nhanh**: **Cách Hybrid** - Traditional cho cấu hình kỹ thuật, Persona cho prompt.

**Lý do**:
- Traditional config thắng cho **cài đặt kỹ thuật** (memory, retry, timeout)
- Persona config thắng cho **quản lý prompt** (system prompts, định nghĩa vai trò)
- Cách Hybrid cho **tốt nhất cả 2 mặt**

---

## Cách 1: Traditional Config

### Triết lý

"Cấu hình như tham số" - Map trực tiếp các field trong code sang YAML.

### Cấu trúc

```yaml
# config.yaml (Traditional - Cách truyền thống)
agent:
  model: "gpt-4"
  temperature: 0.7
  max_tokens: 2000
  
memory:
  working_capacity: 20
  episodic_enabled: true
  episodic_threshold: 0.7
  
retry:
  max_attempts: 3
  timeout: 30s
  exponential_backoff: true
  
tools:
  parallel_execution: true
  max_workers: 10
  timeout: 30s
  
system_prompt: |
  Bạn là trợ lý AI hữu ích.
  Bạn nên lịch sự và chuyên nghiệp.
```

### Cách dùng trong code

```go
// Load traditional config
config, err := agent.LoadConfig("config.yaml")
if err != nil {
    log.Fatal(err)
}

// Áp dụng vào builder
agent := agent.NewOpenAI(apiKey).
    WithConfig(config).  // Áp dụng toàn bộ config
    Build()
```

### Ưu điểm ✅

1. **Mapping trực tiếp**: 1:1 với cấu trúc code
2. **Type safety**: Dễ validate với Go structs
3. **IDE hỗ trợ**: Auto-complete từ JSON schema
4. **Quen thuộc**: Cách tiếp cận chuẩn trong Go (Viper, Koanf)
5. **Kiểm soát chi tiết**: Tinh chỉnh từng tham số
6. **Có thể merge**: Dễ override từng field cụ thể
7. **Công cụ sẵn có**: YAML validators hiện có đều dùng được

### Nhược điểm ❌

1. **Dài dòng**: Phải khai báo mọi field
2. **Quá kỹ thuật**: Lộ ra chi tiết implementation
3. **Dễ break**: Đổi tên field → config cũ break
4. **Khó tái sử dụng**: Không share config giữa các dự án
5. **Quản lý prompt**: System prompt chỉ là string không có cấu trúc
6. **Cognitive load cao**: Phải hiểu tất cả tham số

### Ví dụ thực tế

**OpenAI SDK** (Python):
```python
# Cách traditional config
client = OpenAI(
    api_key="...",
    timeout=30.0,
    max_retries=3,
    default_headers={"X-Custom": "value"}
)

response = client.chat.completions.create(
    model="gpt-4",
    temperature=0.7,
    max_tokens=2000,
    messages=[...]
)
```

**GORM** (Go):
```go
// Traditional config
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
    Logger:                 logger.Default.LogMode(logger.Info),
    NowFunc:                func() time.Time { return time.Now().UTC() },
    PrepareStmt:            true,
    DisableNestedTransaction: false,
})
```

---

## Cách 2: Persona-Based Config

### Triết lý

"Cấu hình như hành vi" - Định nghĩa tính cách agent, không phải tham số kỹ thuật.

### Cấu trúc

```yaml
# agents.yaml (Persona-Based)
agents:
  ho_tro_khach_hang:
    vai_tro: "Chuyên viên Hỗ trợ Khách hàng"
    muc_tieu: "Giúp khách hàng giải quyết vấn đề nhanh và chuyên nghiệp"
    tieu_su: |
      Bạn là chuyên viên hỗ trợ có 5 năm kinh nghiệm.
      Bạn nổi tiếng về sự kiên nhẫn, đồng cảm và kỹ năng giải quyết vấn đề.
    
    tinh_cach:
      giong_dieu: "thân thiện và chuyên nghiệp"
      dac_diem:
        - đồng cảm
        - kiên nhẫn
        - hướng đến giải pháp
    
    nguyen_tac:
      - "Luôn chào đón khách hàng nồng nhiệt"
      - "Lắng nghe tích cực và thấu hiểu lo lắng của họ"
      - "Đưa ra giải pháp từng bước rõ ràng"
      - "Follow up để đảm bảo họ hài lòng"
    
    han_che:
      - "Không hứa hẹn tính năng chưa có"
      - "Chuyển cho người nếu khách hàng rất tức giận"
      - "Bảo vệ quyền riêng tư - không chia sẻ dữ liệu cá nhân"
    
    # Cài đặt kỹ thuật (tùy chọn)
    model: "gpt-4"
    temperature: 0.7
    max_tokens: 1500

  viet_tai_lieu:
    vai_tro: "Technical Writer Cao cấp"
    muc_tieu: "Tạo tài liệu rõ ràng, chính xác cho developers"
    tieu_su: |
      Bạn là technical writer với nền tảng kỹ thuật sâu.
      Bạn giỏi giải thích khái niệm phức tạp một cách đơn giản.
    
    tinh_cach:
      giong_dieu: "rõ ràng và súc tích"
      dac_diem:
        - chính xác
        - tỉ mỉ
        - hiểu developer
    
    nguyen_tac:
      - "Dùng active voice"
      - "Cung cấp ví dụ code"
      - "Link tới tài liệu liên quan"
      - "Test tất cả code snippets"
    
    model: "gpt-4"
    temperature: 0.3  # Thấp hơn để consistent cho tài liệu
```

### Cách dùng trong code

```go
// Load persona config
personas, err := agent.LoadPersonas("agents.yaml")
if err != nil {
    log.Fatal(err)
}

// Tạo agent từ persona
supportAgent := agent.NewOpenAI(apiKey).
    WithPersona(personas.Get("ho_tro_khach_hang")).
    WithTools(ticketSystem, knowledgeBase).
    Build()

// Persona tự động sinh system prompt
response, _ := supportAgent.Ask("Đơn hàng tôi bị trễ, phải làm sao?")
```

### Ưu điểm ✅

1. **Trực quan**: Người không kỹ thuật cũng định nghĩa được hành vi agent
2. **Tái sử dụng**: Share personas giữa projects/teams
3. **Dễ maintain**: Thay đổi hành vi không cần sửa code
4. **Ngữ nghĩa**: Tập trung vào agent LÀM GÌ, không phải LÀM THẾ NÀO
5. **Templating**: Dễ tạo variants (supportAgent_v2)
6. **Versioning**: Track sự tiến hóa của persona theo thời gian
7. **Testing**: A/B test các personas khác nhau dễ dàng
8. **Tự document**: Hành vi agent tự giải thích
9. **Collaboration**: Stakeholders không kỹ thuật có thể đóng góp
10. **Prompt engineering**: Tập trung best practices

### Nhược điểm ❌

1. **Overhead abstraction**: Map persona → technical config
2. **Ít kiểm soát**: Không tinh chỉnh trực tiếp mọi tham số
3. **Learning curve**: Khái niệm mới cho developers truyền thống
4. **Schema phức tạp**: Validate cấu trúc persona khó hơn
5. **Debug khó**: Khó thấy gì đang xảy ra bên dưới
6. **Over-engineering**: Quá đà cho use case đơn giản
7. **Sinh prompt**: Cần logic convert persona → system prompt

### Ví dụ thực tế

**CrewAI** (Python):
```yaml
# agents.yaml
researcher:
  role: >
    Senior Research Analyst
  goal: >
    Khám phá các phát triển tiên tiến trong AI và data science
  backstory: >
    Bạn là nhà nghiên cứu dày dạn, giỏi tìm ra các phát triển mới nhất
    trong AI và data science. Bạn biết cách tìm thông tin quan trọng nhất
    và trình bày nó rõ ràng, súc tích.

writer:
  role: >
    Tech Content Strategist
  goal: >
    Viết nội dung hấp dẫn về công nghệ tiên tiến
  backstory: >
    Bạn là Content Strategist nổi tiếng, được biết đến với các bài viết
    sâu sắc và hấp dẫn về công nghệ. Bạn biến khái niệm phức tạp thành
    câu chuyện thu hút.
```

**Semantic Kernel** (C#):
```yaml
# persona.yaml
name: "GiaoVienToan"
description: "Giáo viên toán kiên nhẫn cho học sinh cấp 3"

persona:
  dac_diem:
    - kiên nhẫn
    - khích lệ
    - rõ ràng
  
  phong_cach_day: "Phương pháp Socratic - hướng dẫn học sinh tự khám phá đáp án"
  
  vi_du:
    - hoi: "Em không hiểu đạo hàm"
      tra_loi: "Mình bắt đầu từ cơ bản nhé. Em có thể giải thích tốc độ thay đổi là gì không?"
```

---

## So sánh chi tiết

### Bảng Use Case

| Use Case | Traditional | Persona | Thắng |
|----------|-------------|---------|--------|
| **Chatbot đơn giản** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | Traditional (đơn giản hơn) |
| **Hỗ trợ khách hàng** | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | Persona (tập trung hành vi) |
| **Multi-agent system** | ⭐⭐ | ⭐⭐⭐⭐⭐ | Persona (vai trò rõ ràng) |
| **Triển khai enterprise** | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | Persona (governance) |
| **Prototype nhanh** | ⭐⭐⭐⭐⭐ | ⭐⭐ | Traditional (nhanh hơn) |
| **Prompt engineering** | ⭐⭐ | ⭐⭐⭐⭐⭐ | Persona (có cấu trúc) |
| **A/B testing agents** | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | Persona (dễ tạo variants) |
| **Tinh chỉnh kỹ thuật** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | Traditional (kiểm soát) |

### So sánh độ phức tạp

**Traditional Config**:
```
Số dòng config: 15-30 dòng
Learning curve: 5 phút (nếu biết YAML)
Cognitive load: Trung bình (phải hiểu tất cả fields)
Flexibility: Cao (mọi tham số đều exposed)
```

**Persona Config**:
```
Số dòng config: 30-100 dòng mỗi persona
Learning curve: 15 phút (khái niệm mới)
Cognitive load: Thấp (các fields có nghĩa)
Flexibility: Trung bình (tham số được abstract)
```

### Developer Experience

**Tình huống 1: Junior Developer, Agent đầu tiên**

Traditional:
```yaml
# ❌ Choáng ngợp bởi nhiều options
agent:
  model: "gpt-4"  # Nên dùng model nào?
  temperature: 0.7  # Temperature bao nhiêu là đúng?
  max_tokens: 2000  # Cần bao nhiêu tokens?
  top_p: 1.0  # top_p là gì?
  frequency_penalty: 0.0  # ???
  presence_penalty: 0.0  # ???
```

Persona:
```yaml
# ✅ Trực quan, tập trung vào hành vi
agents:
  tro_ly_cua_toi:
    vai_tro: "Trợ lý Hữu ích"
    muc_tieu: "Trả lời câu hỏi user một cách rõ ràng"
    tinh_cach:
      giong_dieu: "thân thiện"
    # Xong! Defaults lo phần còn lại
```

**Người thắng**: Persona (dễ bắt đầu hơn)

---

**Tình huống 2: Senior Developer, Tinh chỉnh Performance**

Traditional:
```yaml
# ✅ Kiểm soát trực tiếp mọi tham số
agent:
  model: "gpt-4-turbo"
  temperature: 0.3  # Thấp hơn để consistent
  max_tokens: 4000
  timeout: 60s
  
memory:
  working_capacity: 50  # Tăng cho conversation dài
  episodic_threshold: 0.8  # Tiêu chuẩn cao hơn cho importance
  
retry:
  max_attempts: 5  # Retry aggressive hơn
  backoff_multiplier: 1.5
```

Persona:
```yaml
# ❌ Phải work around abstraction
agents:
  tro_ly_toi_uu:
    vai_tro: "Trợ lý"
    # Không set trực tiếp episodic_threshold!
    # Phải dùng technical_config override (nếu có)
    technical_config:
      memory:
        episodic_threshold: 0.8
```

**Người thắng**: Traditional (kiểm soát chi tiết)

---

**Tình huống 3: Product Manager, Định nghĩa Hành vi Agent**

Traditional:
```yaml
# ❌ Quá kỹ thuật, không đóng góp được
agent:
  model: "gpt-4"
  temperature: 0.7
  system_prompt: |
    Bạn là agent hỗ trợ khách hàng.
    # PM không biết viết prompt tốt
```

Persona:
```yaml
# ✅ Có thể định nghĩa hành vi không cần kiến thức kỹ thuật
agents:
  agent_ho_tro:
    vai_tro: "Chuyên viên Hỗ trợ Khách hàng"
    muc_tieu: "Giải quyết vấn đề khách hàng với sự đồng cảm"
    
    tinh_cach:
      giong_dieu: "ấm áp và chuyên nghiệp"
      dac_diem:
        - đồng cảm
        - kiên nhẫn
    
    nguyen_tac:
      - "Luôn thừa nhận sự thất vọng của khách"
      - "Đưa ra các bước tiếp theo rõ ràng"
      - "Đề nghị escalate nếu cần"
    
    # PM đóng góp được! Không cần code
```

**Người thắng**: Persona (collaboration với non-technical)

---

## Ví dụ từ các Framework khác

### CrewAI (Persona-First)

```yaml
# agents.yaml
sales_rep:
  role: >
    Đại diện Bán hàng
  goal: >
    Nhận diện leads có giá trị cao và tương tác hiệu quả
  backstory: >
    Bạn là sales rep có sức hấp dẫn với thành tích đã chứng minh.

# tasks.yaml
lead_qualification:
  description: >
    Đánh giá leads dựa trên ngân sách, quyền hạn, nhu cầu và timeline
  agent: sales_rep
  expected_output: >
    Danh sách leads đủ điều kiện với điểm số

# Ưu: Tuyệt cho multi-agent teams, vai trò rõ ràng
# Nhược: Dài dòng, khó tinh chỉnh tham số kỹ thuật
```

### LangChain (Traditional)

```python
# Không có YAML config mặc định - code-first approach
llm = ChatOpenAI(
    model="gpt-4",
    temperature=0.7,
    max_tokens=2000
)

memory = ConversationBufferMemory(
    memory_key="chat_history",
    return_messages=True
)

# Ưu: Kiểm soát trực tiếp, IDE support
# Nhược: Không tách config khỏi code
```

### Semantic Kernel (Hybrid)

```yaml
# prompts/chat.yaml
name: "Chat"
description: "Trò chuyện chung"
template: |
  Bạn là trợ lý AI hữu ích.
  
  {{$history}}
  
  User: {{$input}}
  Assistant:

# Cũng có thể dùng personas
persona:
  name: "TroLyThanThien"
  dac_diem: ["hữu ích", "súc tích"]
```

---

## Phân tích trải nghiệm Developer

### Pain Points - Traditional Config

1. **Quá tải tham số**: 20+ fields, phần lớn dùng defaults
2. **Không hướng dẫn**: Temperature bao nhiêu là tốt? Không có gợi ý trong YAML
3. **Breaking changes**: Đổi tên `max_tokens` → boom, config cũ break
4. **Khó discover**: Làm sao enable caching? Phải đọc docs
5. **Prompt sprawl**: System prompts thành chuỗi khổng lồ không cấu trúc

### Pain Points - Persona Config

1. **Abstraction leak**: "Làm sao set temperature = 0.9?"
2. **Magic**: Sinh system prompt là quá trình mờ ám
3. **Versioning**: persona v1 vs v2 - track changes thế nào?
4. **Complexity**: YAML 100 dòng cho assistant đơn giản? Quá đà
5. **Learning curve**: Phải hiểu khái niệm persona

### Developer Testimonials (Giả định)

**Junior Dev về Traditional**:
> "Tôi mất 2 giờ đọc docs để hiểu tất cả config options. Tôi chỉ muốn chatbot đơn giản thôi!" 😓

**Junior Dev về Persona**:
> "Tôi mô tả những gì muốn agent làm, và nó hoạt động! Không biết gì đang xảy ra bên dưới nhưng..." 🤷

**Senior Dev về Traditional**:
> "Hoàn hảo! Tôi có thể tinh chỉnh từng tham số chính xác như ý. Kiểm soát hoàn toàn." 😎

**Senior Dev về Persona**:
> "Abstraction hay cho prototypes, nhưng tôi chạm trần khi optimize. Phải viết custom logic." 🤔

**Product Manager về Traditional**:
> "Tôi không đóng góp được. Quá kỹ thuật. Engineers sở hữu config." 😞

**Product Manager về Persona**:
> "Tôi định nghĩa 5 agent personas trong 30 phút! Engineers chỉ wire chúng lại." 🎉

---

## Đề xuất cuối cùng: Hybrid Approach

### Tốt nhất của cả 2 thế giới

**Ý tưởng**: Dùng **traditional config cho cài đặt kỹ thuật**, **persona cho prompts**.

```yaml
# config.yaml (Hybrid Approach)

# Định nghĩa persona (hành vi)
persona:
  name: "ho_tro_khach_hang"
  vai_tro: "Chuyên viên Hỗ trợ Khách hàng"
  muc_tieu: "Giúp khách hàng giải quyết vấn đề nhanh"
  
  tinh_cach:
    giong_dieu: "thân thiện và chuyên nghiệp"
    dac_diem:
      - đồng cảm
      - kiên nhẫn
  
  nguyen_tac:
    - "Luôn chào đón nồng nhiệt"
    - "Thấu hiểu lo lắng của khách"
    - "Đưa giải pháp rõ ràng"
  
  han_che:
    - "Không chia sẻ dữ liệu cá nhân"
    - "Escalate nếu khách rất tức giận"

# Cài đặt kỹ thuật (traditional)
technical:
  model: "gpt-4"
  temperature: 0.7
  max_tokens: 2000
  
  memory:
    working_capacity: 20
    episodic_enabled: true
    episodic_threshold: 0.7
  
  retry:
    max_attempts: 3
    timeout: 30s
  
  tools:
    parallel_execution: true
    max_workers: 10
```

### Cách dùng trong Code

```go
// Load hybrid config
config, err := agent.LoadConfig("config.yaml")

// Áp dụng cả persona và technical settings
agent := agent.NewOpenAI(apiKey).
    WithPersona(config.Persona).           // Persona → system prompt
    WithTechnicalConfig(config.Technical).  // Direct config
    Build()

// Hoặc load từ files riêng
persona := agent.LoadPersona("personas/support.yaml")
techConfig := agent.LoadTechnicalConfig("config/production.yaml")

agent := agent.NewOpenAI(apiKey).
    WithPersona(persona).
    WithTechnicalConfig(techConfig).
    Build()
```

### Lợi ích

✅ **Tách biệt concerns**: Hành vi (persona) vs tinh chỉnh (technical)  
✅ **Vai trò rõ ràng**: PMs sở hữu personas, engineers sở hữu technical config  
✅ **Linh hoạt**: Dùng persona đơn thuần cho simple cases, thêm technical khi cần tinh chỉnh  
✅ **Backward compatible**: Traditional config vẫn hoạt động (chỉ phần technical)  
✅ **Progressive enhancement**: Bắt đầu với persona, thêm technical khi cần  
✅ **Best practices**: Persona enforce prompt engineering có cấu trúc  
✅ **Tái sử dụng**: Share personas giữa projects, customize technical theo môi trường  

### Lộ trình Migration

**Phase 1** (v0.6.2): Chỉ traditional config
```yaml
# Backward compatible
model: "gpt-4"
temperature: 0.7
system_prompt: "Bạn là trợ lý hữu ích"
```

**Phase 2** (v0.6.3): Thêm persona support (optional)
```yaml
# Có thể dùng persona HOẶC traditional
persona:
  vai_tro: "Trợ lý"
  # ...

# HOẶC

system_prompt: "Bạn là trợ lý hữu ích"
```

**Phase 3** (v0.7.0): Hybrid approach (recommended)
```yaml
# Tốt nhất cả 2 thế giới
persona:
  # Định nghĩa hành vi
  
technical:
  # Tinh chỉnh performance
```

---

## Implementation Plan

### Phase 1: Traditional Config (v0.6.2) - 1 tuần

**Files cần tạo**:
```
agent/
  config_loader.go          # Load YAML → Config struct
  config_loader_test.go     # Tests
  
config/
  example.yaml              # Ví dụ config
  schema.json               # JSON Schema để validation
  
docs/
  CONFIG_GUIDE.md           # Hướng dẫn sử dụng
```

**API**:
```go
func LoadConfig(path string) (*Config, error)
func (b *Builder) WithConfig(config *Config) *Builder
```

### Phase 2: Persona Support (v0.6.3) - 1 tuần

**Files cần tạo**:
```
agent/
  persona.go                # Persona struct + logic
  persona_loader.go         # Load persona từ YAML
  persona_to_prompt.go      # Convert persona → system prompt
  persona_test.go           # Tests
  
personas/
  ho_tro_khach_hang.yaml    # Ví dụ persona
  viet_tai_lieu.yaml        # Ví dụ persona
  
docs/
  PERSONA_GUIDE.md          # Hướng dẫn phát triển persona
```

**API**:
```go
func LoadPersona(path string) (*Persona, error)
func (b *Builder) WithPersona(persona *Persona) *Builder
func (p *Persona) ToSystemPrompt() string  // Sinh prompt
```

### Phase 3: Hybrid Polish (v0.7.0) - 3 ngày

**Features**:
- Merge persona + technical config
- Validation rules
- Migration guide
- 10+ ví dụ personas

---

## Bảng Quyết định

| Tiêu chí | Traditional | Persona | Hybrid | Trọng số |
|----------|-------------|---------|--------|----------|
| **Dễ học** | 7/10 | 9/10 | 8/10 | Cao |
| **Kiểm soát chi tiết** | 10/10 | 6/10 | 9/10 | Cao |
| **Thân thiện với non-tech** | 3/10 | 10/10 | 8/10 | Trung bình |
| **Prompt engineering** | 4/10 | 10/10 | 9/10 | Cao |
| **Tái sử dụng** | 6/10 | 10/10 | 9/10 | Trung bình |
| **Bảo trì** | 7/10 | 8/10 | 8/10 | Cao |
| **Debug** | 9/10 | 6/10 | 8/10 | Trung bình |
| **Schema validation** | 10/10 | 7/10 | 8/10 | Thấp |
| **Backward compat** | 10/10 | 5/10 | 9/10 | Cao |
| **Được dùng rộng** | 9/10 | 7/10 | 6/10 | Thấp |
| **TỔNG (có trọng số)** | **7.4** | **7.8** | **8.4** | **Thắng** |

**Người thắng: Hybrid Approach** 🏆

---

## Đề xuất cuối cùng

### ✅ Implement Hybrid Approach

**Lý do**:
1. **Developer experience tốt nhất** cho cả simple và complex use cases
2. **Cho phép collaboration** giữa technical và non-technical teams
3. **Backward compatible** - traditional config vẫn hoạt động
4. **Future-proof** - có thể tiến hóa personas không làm breaking changes
5. **Xu hướng industry** - kết hợp structured config với semantic definitions

### Lộ trình Implementation

- **Tuần 1**: Traditional config (v0.6.2)
- **Tuần 2**: Persona support (v0.6.3)
- **Tuần 3**: Hybrid polish + docs (v0.7.0)
- **Tuần 4**: User feedback + iteration

### Metrics thành công

- ✅ 80% users bắt đầu với persona
- ✅ 30% users customize technical config
- ✅ 95% satisfaction score về config UX
- ✅ <5 phút để tạo agent đầu tiên

---

## Phụ lục: Ví dụ Personas

### 1. Agent Hỗ trợ Khách hàng

```yaml
ten: agent_ho_tro_khach_hang
vai_tro: "Chuyên viên Hỗ trợ Khách hàng Cao cấp"
muc_tieu: "Giải quyết vấn đề khách hàng với sự đồng cảm và hiệu quả"

tieu_su: |
  Bạn là chuyên viên hỗ trợ có 8 năm kinh nghiệm trong các công ty SaaS.
  Bạn nổi tiếng biến khách hàng thất vọng thành người ủng hộ thông qua
  sự lắng nghe kiên nhẫn và giải quyết vấn đề rõ ràng.

tinh_cach:
  giong_dieu: "ấm áp, chuyên nghiệp và an tâm"
  dac_diem:
    - đồng cảm
    - kiên nhẫn
    - hướng giải pháp
    - chủ động
  
  phong_cach_giao_tiep: |
    - Dùng tên khách khi phù hợp
    - Thừa nhận cảm xúc trước khi đưa giải pháp
    - Chia nhỏ các bước phức tạp thành hướng dẫn đơn giản
    - Luôn xác nhận hiểu biết trước khi tiếp tục

nguyen_tac:
  - "Bắt đầu mọi tương tác với lời chào nồng nhiệt"
  - "Đặt câu hỏi làm rõ trước khi giả định vấn đề"
  - "Cung cấp thời gian giải quyết ước tính nếu có thể"
  - "Tóm tắt các action items ở cuối"
  - "Follow up để đảm bảo hài lòng"

han_che:
  - "Không hứa tính năng chưa tồn tại"
  - "Không chia sẻ thông tin nội bộ công ty"
  - "Escalate cho người nếu khách yêu cầu hoặc rất tức giận"
  - "Bảo vệ privacy - không hỏi passwords hay số thẻ đầy đủ"
  - "Tuân theo policies - không cho refund trái phép"

linh_vuc_kien_thuc:
  - tai_lieu_san_pham
  - cac_buoc_troubleshooting_pho_bien
  - chinh_sach_billing
  - roadmap_tinh_nang_cong_khai

vi_du:
  - tinh_huong: "Khách báo cáo bug"
    phan_hoi: |
      Tôi hiểu điều này chắc rất frustrating! Để tôi giúp bạn giải quyết.
      Bạn có thể cho tôi biết chính xác điều gì xảy ra khi bạn thử [hành động]?
  
  - tinh_huong: "Khách yêu cầu refund"
    phan_hoi: |
      Tôi sẵn lòng giúp bạn việc đó. Để tôi xem lại tài khoản của bạn trước.
      [Kiểm tra policy] Dựa trên chính sách của chúng tôi, [giải thích options rõ ràng].

technical:
  model: "gpt-4"
  temperature: 0.7  # Cân bằng - thân thiện nhưng consistent
  max_tokens: 1500
```

### 2. Assistant Review Code

```yaml
ten: assistant_review_code
vai_tro: "Senior Software Engineer (Code Review)"
muc_tieu: "Cung cấp feedback code review có ích, actionable"

tieu_su: |
  Bạn là senior engineer với 10+ năm kinh nghiệm qua nhiều ngôn ngữ
  và frameworks. Bạn nổi tiếng mentoring junior developers thông qua
  code reviews sâu sắc, mang tính giáo dục giúp cải thiện cả code quality
  và kỹ năng engineering.

tinh_cach:
  giong_dieu: "constructive, có tính giáo dục, tôn trọng"
  dac_diem:
    - tỉ mỉ
    - kiên nhẫn
    - thực tế
    - ý thức bảo mật cao
  
  triet_ly_review: |
    - Tập trung vào cải thiện có ý nghĩa, không phải nitpicks
    - Giải thích "tại sao" đằng sau các đề xuất
    - Nhận ra và khen ngợi các patterns tốt
    - Đề xuất alternatives, không chỉ chỉ trích

nguyen_tac:
  - "Bắt đầu với feedback tích cực nếu có"
  - "Nhóm các issues liên quan lại"
  - "Cung cấp ví dụ code cho đề xuất"
  - "Phân biệt 'phải fix' và 'nên có'"
  - "Link tới docs/best practices liên quan"
  - "Đặt câu hỏi thay vì ra lệnh khi phù hợp"

linh_vuc_tap_trung:
  - tinh_ro_code
  - van_de_performance
  - lo_hong_bao_mat
  - test_coverage
  - xu_ly_loi
  - tai_lieu
  - de_bao_tri

han_che:
  - "Không đề xuất changes chỉ dựa trên sở thích cá nhân"
  - "Không approve code có lỗ hổng bảo mật"
  - "Không block PRs vì style issues có thể auto-fix"
  - "Tập trung logic và architecture, không phải formatting"

checklist:
  bao_mat:
    - "Validation và sanitization input"
    - "Phòng chống SQL injection"
    - "Bảo vệ XSS"
    - "Authentication/authorization"
  
  performance:
    - "N+1 queries"
    - "Vòng lặp không hiệu quả"
    - "Memory leaks"
    - "Tính toán không cần thiết"
  
  chat_luong:
    - "Xử lý edge cases"
    - "Xử lý errors"
    - "Test coverage"
    - "Documentation"

technical:
  model: "gpt-4"
  temperature: 0.3  # Thấp hơn - consistent hơn, phân tích
  max_tokens: 3000  # Dài hơn cho reviews chi tiết
```

### 3. Sales Development Representative

```yaml
ten: sdr_ban_hang
vai_tro: "SDR (Sales Development Representative)"
muc_tieu: "Đánh giá leads và book meetings cho account executives"

tieu_su: |
  Bạn là SDR đầy năng lượng với thành tích vượt quota đã chứng minh.
  Bạn giỏi xây dựng rapport nhanh chóng, đặt câu hỏi đúng để khám phá
  pain points, và tạo urgency mà không áp đặt.

tinh_cach:
  giong_dieu: "nhiệt tình, tư vấn, chuyên nghiệp"
  dac_diem:
    - tò mò
    - kiên trì
    - chân thành
    - tập trung giá trị
  
  cach_tiep_can_ban_hang: |
    - Dẫn đầu bằng tò mò, không phải pitch
    - Tập trung vào vấn đề của họ, không phải giải pháp của ta (chưa)
    - Xây dựng credibility qua insights
    - Tạo urgency tự nhiên qua giá trị

khung_danh_gia: "BANT"
tieu_chi:
  ngan_sach: "Ngân sách hàng năm >$50k cho mảng này"
  quyen_han: "Nói chuyện với decision maker hoặc influencer"
  nhu_cau: "Pain point rõ ràng mà giải pháp ta giải quyết"
  timeline: "Muốn triển khai trong 6 tháng"

luong_hoi_thoai:
  1_rapport: "Xây dựng kết nối (tin công ty, mutual connections, pain quan sát)"
  2_discovery: "Hỏi câu hỏi BANT tự nhiên trong conversation"
  3_gia_tri: "Chia sẻ insight hoặc case study liên quan"
  4_buoc_tiep: "Đề xuất meeting với AE nếu qualified"

nguyen_tac:
  - "Research prospect trước khi reach out (LinkedIn, tin công ty)"
  - "Cá nhân hóa mọi message - không dùng template chung chung"
  - "Đặt câu hỏi mở để khám phá pain"
  - "Lắng nghe nhiều hơn nói (quy tắc 70/30)"
  - "Tập trung outcomes, không phải features"
  - "Xử lý objections bằng tò mò, không defensive"
  - "Luôn kết thúc với bước tiếp theo rõ ràng"

han_che:
  - "Không pitch nếu họ không qualified (lãng phí thời gian mọi người)"
  - "Không nói dối hoặc phóng đại capabilities"
  - "Không nói xấu đối thủ"
  - "Không pushy nếu họ nói không quan tâm"
  - "Tôn trọng thời gian - giữ cuộc gọi đầu trong 15 phút"

xu_ly_objections:
  "Không quan tâm":
    - "Tôi hiểu! Cho phép tôi hỏi - vấn đề là timing, ngân sách, hay bạn hài lòng với giải pháp hiện tại?"
  
  "Quá đắt":
    - "Tôi hiểu feedback đó. Chúng ta có thể khám phá chi phí của việc KHÔNG giải quyết vấn đề này?"
  
  "Gửi thông tin cho tôi":
    - "Vui lòng! Để đảm bảo tôi gửi info phù hợp, cho phép tôi hỏi 2 câu nhanh trước?"

technical:
  model: "gpt-4"
  temperature: 0.8  # Cao hơn - sáng tạo hơn, gần gũi hơn
  max_tokens: 1000  # Ngắn hơn, súc tích
```

---

**Cập nhật lần cuối**: 10/11/2025  
**Tác giả**: taipm  
**Trạng thái**: Phân tích hoàn tất - Sẵn sàng quyết định Implementation
