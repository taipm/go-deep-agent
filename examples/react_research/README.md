# Multi-Step Research Example

Demonstrates complex ReAct flow with multiple tools and information gathering.

## Features

- **2 Tools**: search (mock) + summarize
- **Multi-step reasoning**: 3-5 iterations per task
- **Tool chaining**: Search → process → summarize
- **Step tracking**: Full trace of tool calls

## Tools

### search
Mock knowledge base with predefined answers for:
- golang, react, python, kubernetes
- ai, llm, gpt, agent

### summarize
Text summarization with word count.

## Usage

```bash
export OPENAI_API_KEY="your-key"
go run main.go
```

## Expected Flow

**Task**: "What is Go and who created it?"

1. **Thought**: Need to search for Go information
2. **Action**: search(query="golang")
3. **Observation**: "Go is a statically typed..."
4. **Thought**: Got answer, can respond
5. **Final Answer**: Formatted response

**Task**: "Compare AI agents and LLMs"

1. **Thought**: Need info on both topics
2. **Action**: search(query="ai agent")
3. **Observation**: "AI agents are autonomous..."
4. **Thought**: Now need LLM info
5. **Action**: search(query="llm")
6. **Observation**: "Large Language Models..."
7. **Thought**: Have both, need to summarize
8. **Action**: summarize(text="...")
9. **Observation**: Summary result
10. **Final Answer**: Comparison with summary

## Output

```
Multi-Step Research Example
============================

Task 1: What is Go programming language and who created it?
------------------------------------------------------------

Final Answer:
Go is a statically typed, compiled language created at Google in 2009...

Stats:
  Iterations: 2
  Steps: 4
  
Tool Calls: 1
Tool Sequence:
  1. search
```
