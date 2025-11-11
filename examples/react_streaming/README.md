# Streaming ReAct Example

Real-time streaming of ReAct execution events.

## Features

- **Event Stream**: Real-time updates via channels
- **Progress Tracking**: Live iteration and step counts
- **Event Types**: start, thought, action, observation, final, error, complete
- **Async Tools**: Weather and news with simulated delays

## Usage

```bash
export OPENAI_API_KEY="your-key"
go run main.go
```

## Output

```
Streaming ReAct Example
=======================

Task: Get the weather in Paris and latest tech news

🚀 Starting ReAct execution...

💭 [Iteration 1] THOUGHT:
   I need to get weather for Paris and tech news...

🔧 [Step 1] ACTION:
   Tool: weather
👁️  OBSERVATION:
   Paris: 22°C, Sunny

💭 [Iteration 2] THOUGHT:
   Now I need to get tech news...

🔧 [Step 2] ACTION:
   Tool: news
👁️  OBSERVATION:
   Latest on tech: Breaking news update

💭 [Iteration 3] THOUGHT:
   I have both pieces of information...

✅ FINAL ANSWER:
   The weather in Paris is 22°C and sunny. Latest tech news: Breaking news...

📊 Execution Complete!
   Total iterations: 3
   Total steps: 2

Done!
```

## Event Types

| Event | Description |
|-------|-------------|
| `start` | Execution begins |
| `thought` | LLM reasoning step |
| `action` | Tool call initiated |
| `observation` | Tool result received |
| `final` | Final answer ready |
| `error` | Error occurred |
| `complete` | Execution finished |

## Code Pattern

```go
events, err := ai.StreamReAct(ctx, task)
if err != nil {
    log.Fatal(err)
}

for event := range events {
    switch event.Type {
    case "thought":
        fmt.Printf("Thinking: %s\n", event.Content)
    case "action":
        fmt.Printf("Tool: %s\n", event.Step.Tool)
    case "observation":
        fmt.Printf("Result: %s\n", event.Content)
    case "final":
        fmt.Printf("Answer: %s\n", event.Content)
    }
}
```

## Benefits

- **Responsive UX**: Show progress to users
- **Debugging**: Watch agent reasoning in real-time
- **Monitoring**: Track iterations and tool usage
- **Early termination**: Cancel via context
