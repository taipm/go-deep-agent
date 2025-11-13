# LLM Adapters

This package contains provider-specific implementations of the `LLMAdapter` interface, enabling seamless integration with multiple LLM providers.

## Architecture

```
agent/
├── adapter.go           # LLMAdapter interface and unified data structures
└── adapters/
    ├── README.md        # This file
    ├── gemini_adapter.go      # Google Gemini implementation
    └── gemini_adapter_test.go # Unit tests
```

## Available Adapters

### Google Gemini

**File:** `gemini_adapter.go`  
**Provider:** Google Generative AI  
**Models:** `gemini-pro`, `gemini-pro-vision`, etc.

**Key Features:**
- Temperature clamping to Gemini's 0-1 range
- System prompt via `SystemInstruction` (not message)
- Role mapping: "assistant" → "model"
- Parts-based content structure
- Iterator-based streaming

**Usage:**
```go
adapter, err := adapters.NewGeminiAdapter("your-api-key")
if err != nil {
    log.Fatal(err)
}
defer adapter.Close()

resp, err := adapter.Complete(ctx, &agent.CompletionRequest{
    Model: "gemini-pro",
    Messages: []agent.Message{
        {Role: "user", Content: "Hello!"},
    },
})
```

## Testing

### Run Unit Tests

```bash
# Run all tests
go test ./agent/adapters/

# Run with verbose output
go test -v ./agent/adapters/

# Run in short mode (skip tests requiring API client)
go test -short ./agent/adapters/

# Generate coverage report
go test -coverprofile=coverage.out ./agent/adapters/
go tool cover -html=coverage.out -o coverage.html
```

### Test Coverage

**Current Coverage:**
- Constructor validation: 100%
- Temperature clamping: 100%
- Message conversion: 100%
- Tool conversion: 100%
- Callback safety: 100%
- Edge cases: 100%

**Total:** 18 unit test cases (16 passing, 2 skipped)

### Integration Tests

Integration tests require real API keys and use the `integration` build tag to skip by default.

**Available Integration Tests:**

**Gemini Adapter:** `gemini_adapter_integration_test.go`
- Complete() with real API (simple, system prompt, temperature, history)
- Stream() with real API (with/without callback)
- Context cancellation and timeout
- Error handling (invalid model, empty messages)
- MaxTokens parameter validation
- Stop sequences
- Concurrent requests

**Setup:**

1. Set environment variable:
   ```bash
   export GEMINI_API_KEY="your-api-key-here"
   ```

2. Run integration tests:
   ```bash
   # Run all integration tests
   go test -tags=integration -v ./agent/adapters/
   
   # Run specific integration test
   go test -tags=integration -v -run TestIntegrationGeminiAdapterComplete ./agent/adapters/
   
   # Run with timeout
   go test -tags=integration -v -timeout 5m ./agent/adapters/
   ```

**Note:** Integration tests make real API calls and may incur costs. They are skipped in normal test runs.

## Adding New Adapters

To add support for a new LLM provider:

### 1. Create Adapter File

```go
// adapters/provider_adapter.go
package adapters

import (
    "context"
    "github.com/taipm/go-deep-agent/agent"
)

type ProviderAdapter struct {
    client *provider.Client
}

func NewProviderAdapter(apiKey string) (*ProviderAdapter, error) {
    // Initialize provider client
}

func (a *ProviderAdapter) Complete(ctx context.Context, req *agent.CompletionRequest) (*agent.CompletionResponse, error) {
    // Convert req to provider format
    // Call provider API
    // Convert response to unified format
}

func (a *ProviderAdapter) Stream(ctx context.Context, req *agent.CompletionRequest, onChunk func(string)) (*agent.CompletionResponse, error) {
    // Implement streaming
}
```

### 2. Handle Provider-Specific Quirks

Each provider has unique requirements:

**Temperature Range:**
- OpenAI: 0-2
- Gemini: 0-1 (clamp if needed)
- Anthropic: 0-1

**System Prompt:**
- OpenAI: System message in messages array
- Gemini: SystemInstruction parameter
- Anthropic: Separate system parameter

**Role Names:**
- OpenAI: "assistant"
- Gemini: "model"
- Anthropic: "assistant"

**Content Structure:**
- OpenAI: Simple string
- Gemini: Parts array
- Anthropic: Content blocks

### 3. Create Test File

Follow the pattern in `gemini_adapter_test.go`:

```go
// adapters/provider_adapter_test.go
package adapters

import "testing"

const testModel = "provider-model"

func TestNewProviderAdapter(t *testing.T) {
    tests := []struct {
        name    string
        apiKey  string
        wantErr bool
    }{
        {
            name:    "[P1] valid API key",
            apiKey:  "test-key",
            wantErr: false,
        },
        // ...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}

// Add tests for:
// - Temperature clamping (if needed)
// - Message conversion
// - Tool conversion
// - Response conversion
// - Edge cases
```

### 4. Test Checklist

- [ ] Constructor with valid/invalid API keys
- [ ] Temperature clamping (if range differs from standard)
- [ ] Message format conversion
- [ ] Tool format conversion
- [ ] Response format conversion
- [ ] Edge cases (empty inputs, long messages, special characters)
- [ ] Callback safety (nil callbacks)
- [ ] Resource cleanup (Close method)

## Design Principles

### Keep Adapters Thin

Each adapter should be ~150-250 lines of conversion logic:
- ✅ Convert unified format → provider format
- ✅ Call provider SDK
- ✅ Convert provider response → unified format
- ❌ Don't add business logic here
- ❌ Don't handle caching/rate limiting here (Builder's job)

### Handle Provider Quirks Internally

All provider-specific behavior should be contained in the adapter:
- Temperature range clamping
- System prompt handling
- Role name mapping
- Content structure conversion
- Streaming implementation differences

The rest of the codebase should never know which provider is being used.

### Fail Fast

Validate inputs early and return clear errors:
```go
if req == nil {
    return nil, fmt.Errorf("request cannot be nil")
}
if req.Model == "" {
    return nil, fmt.Errorf("model is required")
}
```

### Document Differences

Use comments to explain provider-specific behavior:
```go
// Gemini only supports temperature 0-1, clamp if needed
if temp > 1.0 {
    temp = 1.0
}
```

## Testing Best Practices

### Use Table-Driven Tests

```go
tests := []struct {
    name    string
    input   interface{}
    want    interface{}
    wantErr bool
}{
    {name: "case 1", input: x, want: y},
    {name: "case 2", input: a, want: b},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        // Test logic
    })
}
```

### Tag Tests with Priority

- `[P1]` - High priority (critical paths)
- `[P2]` - Medium priority (edge cases)
- `[P3]` - Low priority (nice-to-have)

```go
{name: "[P1] valid input", ...},
{name: "[P2] empty input", ...},
```

### Skip Tests Requiring API Client

```go
if testing.Short() {
    t.Skip("Skipping test that requires API client")
}
```

Run with `-short` flag to skip: `go test -short ./...`

### Use Constants for Repeated Values

```go
const testModel = "gemini-pro"

// Use testModel instead of hardcoding
req := &agent.CompletionRequest{
    Model: testModel,
    // ...
}
```

## Resources

- [LLMAdapter Interface](../adapter.go) - Interface definition and documentation
- [Test Automation Guide](../../docs/test-automation-gemini-adapter.md) - Detailed testing guide
- [Contributing Guide](../../CONTRIBUTING.md) - Project contribution guidelines

## Questions?

For questions about:
- **Interface design:** See `agent/adapter.go`
- **Testing patterns:** See `gemini_adapter_test.go`
- **Provider specifics:** Check provider's official documentation
