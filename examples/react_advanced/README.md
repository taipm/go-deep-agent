# Advanced Configuration Example

Demonstrates ReAct advanced features: custom templates, few-shot examples, and callbacks.

## Features

- **Custom Prompt Template**: Tailored ReAct instructions
- **Few-shot Examples**: Pre-loaded example for guidance
- **Enhanced Callbacks**: Real-time step notifications with emojis
- **Callback tracking**: OnThought, OnAction, OnObservation, OnFinal, OnCompleted

## Usage

```bash
export OPENAI_API_KEY="your-key"
go run main.go
```

## Output

```
Advanced Configuration Example
================================

Task: Search for information about Kubernetes
----------------------------------------------------------------------
💭 [1] Thinking: I need to search for information about Kubernetes...
🔧 [1] Tool: search
👁️  [1] Result: Found: Kubernetes is a container orchestration platform...
💭 [2] Thinking: I have the information needed...
✅ [2] Answer ready

📊 Execution complete: 2 iterations

Final Answer:
Kubernetes is a container orchestration platform...

Metrics:
  Iterations: 2
  Steps: 4
```

## Configuration Details

### Custom Template
```go
customTemplate := `You are a research assistant using ReAct.

Available Tools:
{tools}

Examples:
{examples}

Rules:
- Always THOUGHT before ACTION
- Use tools for information
- Provide detailed FINAL answers

Task: {task}`
```

### Few-shot Example
```go
examples := []agent.ReActExample{
    {
        Task: "Find info about Go",
        Steps: []string{
            `THOUGHT: Need to search for Go`,
            `ACTION: search(query="golang")`,
            `OBSERVATION: Go is a programming language`,
            `FINAL: Go is a statically typed language`,
        },
    },
}
```

### Enhanced Callback
```go
callback := &agent.EnhancedReActCallback{
    OnThought: func(content string, iteration int) {
        fmt.Printf("💭 [%d] Thinking: %s\n", iteration, content)
    },
    OnAction: func(tool string, args map[string]interface{}, iteration int) {
        fmt.Printf("🔧 [%d] Tool: %s\n", iteration, tool)
    },
    OnObservation: func(content string, iteration int) {
        fmt.Printf("👁️ [%d] Result: %s\n", iteration, content)
    },
    OnFinal: func(answer string, iteration int) {
        fmt.Printf("✅ [%d] Answer ready\n", iteration)
    },
    OnCompleted: func(result *agent.ReActResult) {
        fmt.Printf("Complete: %d iterations\n", result.Iterations)
    },
}
```
