# Simple ReAct Example

Basic calculator demonstrating ReAct pattern with tool execution.

## Features

- Single tool: arithmetic calculator
- Simple task execution
- Iteration tracking
- Error handling

## Usage

```bash
export OPENAI_API_KEY="your-key"
go run main.go
```

## Expected Output

```
Simple ReAct Example
====================

Task 1: What is 25 + 17?
Answer: 42.00
Stats: 1 iterations

Task 2: Calculate 100 divided by 4
Answer: 25.00
Stats: 1 iterations

Done!
```

## ReAct Flow

1. **Thought**: LLM analyzes task
2. **Action**: Calls calculator tool
3. **Observation**: Gets result
4. **Final Answer**: Returns formatted answer
