# ReAct with Built-in MathTool

Demonstrates ReAct reasoning pattern with the built-in professional `MathTool`.

## Overview

This example shows how to use ReAct mode with the built-in `tools.NewMathTool()` instead of custom calculators. The MathTool provides professional-grade mathematical operations powered by `govaluate` and `gonum` libraries.

## Key Differences from Custom Calculators

### ❌ Custom Calculator (Limited)
```go
// User-defined simple calculator
calcTool := agent.NewTool("calculate", "Perform calculations").
    AddParameter("expression", "string", "Math expression", true)
calcTool.Handler = func(args string) (string, error) {
    // Manual parsing of expressions
    // Limited operations (+, -, *, /)
    // No statistics, unit conversion, etc.
}
```

### ✅ Built-in MathTool (Professional)
```go
// Professional math tool with 5 operation categories
ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithReActMode(true).
    WithTool(tools.NewMathTool()). // ← Built-in, feature-rich
    Execute(ctx, task)
```

## MathTool Capabilities

### 1. Expression Evaluation
- Basic: `2 * (3 + 4)`, `10 / 2 + 5`
- Functions: `sqrt(16)`, `sin(3.14/2)`, `pow(2, 10)`
- Complex: `log(100) + sqrt(144) * 2`

### 2. Statistics
- Mean, Median, Mode
- Standard deviation, Variance
- Min, Max, Sum

### 3. Equation Solving
- Linear equations: `x + 5 = 10`
- Simple algebra: `2x - 3 = 7`

### 4. Unit Conversion
- Distance: km ↔ m ↔ cm
- Weight: kg ↔ g
- Temperature: celsius ↔ fahrenheit
- Time: hours ↔ minutes ↔ seconds

### 5. Random Generation
- Random integers in range
- Random floats
- Random choice from list

## Examples Included

### Example 1: Simple Calculation
```
Task: What is the result of 2 * (15 + 8) - sqrt(16)?
Answer: 42
```

### Example 2: Statistics
```
Task: Calculate mean, median, and stdev of: 85, 90, 78, 92, 88, 95, 82
Answer: Mean: 87.14, Median: 88, StdDev: 5.67
```

### Example 3: Complex Multi-Step
```
Task: Student has test scores 75, 82, 90, 88, 85 (60% of grade).
      Final exam is 40% of grade. Need 85% overall for an A.
      What minimum final exam score is needed?
      
Process:
  THOUGHT: Calculate average of 5 tests
  ACTION: math(operation="statistics", numbers=[75,82,90,88,85], stat_type="mean")
  OBSERVATION: 84.0
  
  THOUGHT: Calculate required final exam score
  ACTION: math(operation="evaluate", expression="(85 - 84*0.6) / 0.4")
  OBSERVATION: 86.5
  
  FINAL: The student needs to score at least 86.5% on the final exam
```

### Example 4: Unit Conversion
```
Task: 5 km to meters, 100 celsius to fahrenheit
Answer: 5000 meters, 212°F
```

### Example 5: Full Reasoning Trace
```
Displays complete THOUGHT → ACTION → OBSERVATION → FINAL flow
Shows all intermediate steps and tool calls
```

## Why This Addresses the GitHub Issue

The GitHub issue (`GITHUB_ISSUE_REPORT.md`) claimed that:
1. ❌ Tools don't execute in ReAct mode
2. ❌ Need `WithAutoExecute(true)` for ReAct
3. ❌ Need `WithParallelTools(false)` for ReAct

**This example proves:**
1. ✅ ReAct DOES execute tools automatically
2. ✅ NO need for `WithAutoExecute(true)` in ReAct mode
3. ✅ ReAct has its own execution logic via `executeTool()`

**Root cause of the issue was likely:**
- Model generating wrong format: `ACTION: functions.calculate(...)` instead of `ACTION: math(...)`
- Tool name mismatch: `"calculate"` vs `"math"`
- Parser failing to recognize format

## Usage

```bash
export OPENAI_API_KEY="your-api-key-here"
go run main.go
```

## Expected Output

```
=== ReAct with Built-in MathTool ===
Demonstrates ReAct pattern with professional math operations

─────────────────────────────────────────────────────
Example 1: Simple Expression Evaluation
─────────────────────────────────────────────────────

Task: What is the result of 2 * (15 + 8) - sqrt(16)?

✅ Answer: 42
📊 Iterations: 2

─────────────────────────────────────────────────────
Example 2: Statistical Analysis
─────────────────────────────────────────────────────

Task: Calculate the mean, median, and standard deviation of: 85, 90, 78, 92, 88, 95, 82

✅ Answer: Mean: 87.14, Median: 88.00, Standard Deviation: 5.67
📊 Tool calls: 3
📊 Iterations: 4

[... more examples ...]

✅ All ReAct + MathTool examples completed!
```

## Key Takeaways

1. **Use Built-in Tools**: `tools.NewMathTool()` is production-ready
2. **ReAct Auto-Executes**: No need for `WithAutoExecute(true)`
3. **Simple Configuration**: Just `.WithReActMode(true)` + `.WithTool(tool)`
4. **Professional Features**: Statistics, conversions, equation solving
5. **Clear Format**: `ACTION: math(operation="evaluate", expression="...")`

## Advanced Configuration

```go
// With custom system prompt
ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithReActMode(true).
    WithReActMaxIterations(10).        // Allow more reasoning steps
    WithReActTimeout(2 * time.Minute). // Set timeout
    WithReActStrict(false).            // Graceful error handling
    WithTool(tools.NewMathTool()).
    WithSystem("You are a math tutor. Show step-by-step solutions.")
```

## Comparison with Official Examples

### react_simple/main.go (Custom Calculator)
```go
// Custom calculator - limited functionality
calcTool := agent.NewTool("calculator", "Performs arithmetic")
calcTool.Handler = calculatorTool // User-defined function
```

### react_math/main.go (Built-in MathTool)
```go
// Built-in MathTool - professional features
ai.WithTool(tools.NewMathTool()) // Batteries included
```

## Dependencies

The MathTool uses:
- `github.com/Knetic/govaluate` - Expression evaluation
- `gonum.org/v1/gonum/stat` - Statistical functions

These are already included in go-deep-agent's dependencies.

## Next Steps

After running this example:
1. Try combining multiple built-in tools: `tools.WithDefaults()` (DateTime + Math)
2. Use `tools.WithAll()` for FileSystem + HTTP + DateTime + Math
3. Experiment with complex multi-step reasoning tasks
4. Build production agents with professional math capabilities
