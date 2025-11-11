# GitHub Issue Report - ReAct Tool Execution Problem

## Issue Title
🐛 ReAct mode with tools: Tools not executing, WithDefaults() doesn't enable required settings

---

## Summary

When using `WithReActMode(true)` with custom tools, the tools are **not executed** even though the model generates correct ACTION statements. This appears to be due to missing default configurations that should be automatically enabled.

---

## Environment

- **Library Version:** go-deep-agent v0.7.2
- **Go Version:** go1.25.2
- **OS:** macOS (Darwin 25.1.0)
- **Model:** gpt-4o-mini, gpt-4o
- **OpenAI API:** Latest

---

## Expected Behavior

When using ReAct mode with tools:

```go
agent := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithDefaults().
    WithReActMode(true).
    WithReActMaxIterations(5).
    WithTool(mathTool)

result, err := agent.Execute(ctx, prompt)
```

**Expected ReAct Loop:**
```
THOUGHT: I need to verify the calculation
ACTION: functions.calculate(expression="500000*0.5")
OBSERVATION: 250000.00                              ← Tool should execute!
THOUGHT: Now I need to add...
ACTION: functions.calculate(expression="500000+250000")
OBSERVATION: 750000.00                              ← Tool should execute!
THOUGHT: The calculation is correct
FINAL ANSWER: Document verified
```

---

## Actual Behavior

Tools are **NOT executed**:

```
THOUGHT: I need to verify the calculation
ACTION: functions.calculate(expression="500000*0.5")
FINAL: [Waiting for the tool's response]           ← Tool NOT executed!
```

**The tool handler is never called, and no OBSERVATION is generated.**

---

## Reproduction Code

### Complete Minimal Example

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "os"
    "strconv"
    "strings"

    "github.com/taipm/go-deep-agent/agent"
)

// Simple math tool
func mathToolHandler(argsJSON string) (string, error) {
    var args map[string]interface{}
    if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
        return "", err
    }

    expr, ok := args["expression"].(string)
    if !ok {
        return "", fmt.Errorf("expression must be a string")
    }

    // Simple calculator for demonstration
    expr = strings.ReplaceAll(expr, " ", "")
    parts := strings.Split(expr, "+")
    if len(parts) == 2 {
        a, _ := strconv.ParseFloat(parts[0], 64)
        b, _ := strconv.ParseFloat(parts[1], 64)
        return fmt.Sprintf("%.2f", a+b), nil
    }

    return "", fmt.Errorf("unsupported expression")
}

func main() {
    apiKey := os.Getenv("OPENAI_API_KEY")
    if apiKey == "" {
        log.Fatal("OPENAI_API_KEY required")
    }

    // Create math tool
    mathTool := agent.NewTool("calculate", "Perform calculations").
        AddParameter("expression", "string", "Math expression", true)
    mathTool.Handler = mathToolHandler

    // Create agent with ReAct mode
    reviewer := agent.NewOpenAI("gpt-4o-mini", apiKey).
        WithDefaults().
        WithReActMode(true).
        WithReActMaxIterations(5).
        WithTool(mathTool).
        WithSystem("Verify calculations using the calculate tool")

    ctx := context.Background()
    result, err := reviewer.Execute(ctx, "Calculate 100 + 200")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Answer: %s\n", result.Answer)
    fmt.Printf("Steps:\n")
    for _, step := range result.Steps {
        fmt.Printf("  [%s] %s\n", step.Type, step.Content)
    }
}
```

### Output (Problem)

```
THOUGHT: I need to calculate 100 + 200
ACTION: functions.calculate(expression="100+200")
FINAL: [Waiting for the tool's response]

Answer: Waiting for the tool's response
Steps:
  [thought] I need to calculate 100 + 200
  [action] functions.calculate(expression="100+200")
  [final] Waiting for the tool's response
```

**Tool handler is NEVER called!**

---

## Root Cause Analysis

### Problem 1: WithDefaults() doesn't enable tool execution

When using ReAct mode with tools, these settings are **required** but NOT enabled by `WithDefaults()`:

```go
.WithAutoExecute(true)       // Required for tool execution
.WithParallelTools(false)    // Required for sequential ReAct pattern
```

### Problem 2: Model uses parallel tools by default

OpenAI models try to optimize by using `multi_tool_use.parallel`:

```
ACTION: multi_tool_use.parallel(
  tool_uses=[
    {"recipient_name": "functions.calculate", ...},
    {"recipient_name": "functions.calculate", ...}
  ]
)
```

This conflicts with ReAct's sequential THOUGHT-ACTION-OBSERVATION pattern.

### Problem 3: No error or warning

The library **silently fails** without any error message or warning that tools are not being executed.

---

## Workaround (Current)

Must manually add missing configurations:

```go
reviewer := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithDefaults().
    WithReActMode(true).
    WithReActMaxIterations(5).
    WithAutoExecute(true).        // ← Must add manually!
    WithParallelTools(false).     // ← Must add manually!
    WithTool(mathTool)
```

**But this still doesn't work!** Tools are still not executed.

---

## Proposed Solutions

### Solution 1: Auto-enable required settings in WithReActMode()

When `WithReActMode(true)` is called, automatically enable:

```go
func (b *AgentBuilder) WithReActMode(enable bool) *AgentBuilder {
    b.reActMode = enable
    if enable {
        // Auto-enable required settings for ReAct
        b.autoExecute = true          // Tools must execute
        b.parallelTools = false       // Sequential only
    }
    return b
}
```

**Benefit:** Developer Experience - just works out of the box!

---

### Solution 2: Include in WithDefaults()

`WithDefaults()` should enable settings needed for common use cases:

```go
func (b *AgentBuilder) WithDefaults() *AgentBuilder {
    return b.
        WithMemory().
        WithAutoExecute(true).        // ← Add this
        WithParallelTools(false).     // ← Add this
        // ... other defaults
}
```

**Benefit:** One call covers most use cases

---

### Solution 3: Add validation and warnings

Detect incompatible configurations and warn:

```go
func (b *AgentBuilder) Build() (*Agent, error) {
    // Validate ReAct + Tools configuration
    if b.reActMode && len(b.tools) > 0 {
        if !b.autoExecute {
            return nil, fmt.Errorf(
                "ReAct mode with tools requires WithAutoExecute(true)")
        }
        if b.parallelTools {
            log.Warn("ReAct mode works best with WithParallelTools(false)")
        }
    }
    return agent, nil
}
```

**Benefit:** Clear error messages guide developers

---

### Solution 4: Better documentation

Add clear documentation about required configurations:

```go
// WithReActMode enables ReAct reasoning pattern.
//
// IMPORTANT: When using tools with ReAct mode, you must also call:
//   - WithAutoExecute(true) - to enable automatic tool execution
//   - WithParallelTools(false) - for sequential execution
//
// Example:
//   agent.NewOpenAI(model, key).
//       WithReActMode(true).
//       WithAutoExecute(true).        // Required!
//       WithParallelTools(false).     // Required!
//       WithTool(myTool)
func (b *AgentBuilder) WithReActMode(enable bool) *AgentBuilder
```

---

## Impact

### Developer Experience Issues

1. **Not intuitive** - `WithDefaults()` doesn't include required settings
2. **Silent failure** - No error when tools aren't configured correctly
3. **Hidden requirements** - Developers must discover `WithAutoExecute()` and `WithParallelTools()` through trial and error
4. **Inconsistent with examples** - Official examples don't show these settings

### Affected Use Cases

Any use case combining:
- ✅ ReAct mode
- ✅ Custom tools
- ❌ **Doesn't work without manual configuration**

**Common scenarios:**
- Math verification agents
- Research agents with search tools
- Code review agents with analysis tools
- Data processing agents with transformation tools

---

## Comparison with Official Examples

Your official examples (`/examples/react_simple/main.go`, `/examples/react_advanced/main.go`) show:

```go
ai := agent.NewOpenAI("gpt-4o-mini", apiKey).
    WithReActMode(true).
    WithReActMaxIterations(3).
    WithTool(calcTool)
// No WithAutoExecute!
// No WithParallelTools!
```

This is **confusing** because:
1. Examples don't show required settings
2. Examples may not actually work in all cases
3. Developers copy examples and encounter issues

---

## Questions

1. **Is tool execution automatic in ReAct mode?**
   - If yes, why isn't it working?
   - If no, why don't examples show `WithAutoExecute(true)`?

2. **Should `WithDefaults()` include `WithAutoExecute(true)`?**
   - Most use cases with tools need this
   - Would improve developer experience

3. **Should `WithReActMode(true)` auto-enable required settings?**
   - Would prevent configuration errors
   - Would make API more intuitive

4. **Is this a known issue in v0.7.2?**
   - Should we upgrade/downgrade?
   - Is there a workaround?

---

## Suggestions for API Improvement

### Make ReAct "just work" with tools

```go
// Current (confusing)
agent.NewOpenAI(model, key).
    WithDefaults().
    WithReActMode(true).
    WithAutoExecute(true).        // Why not automatic?
    WithParallelTools(false).     // Why not automatic?
    WithTool(tool)

// Proposed (intuitive)
agent.NewOpenAI(model, key).
    WithDefaults().
    WithReActMode(true).          // Auto-enables required settings!
    WithTool(tool)                // Just works!
```

### Add convenience method

```go
// Convenience method for common case
agent.NewOpenAI(model, key).
    WithReActAndTools(tool1, tool2).  // All settings included!
    WithReActMaxIterations(5)
```

### Improve error messages

```go
// If misconfigured
Error: ReAct mode with tools requires WithAutoExecute(true).
Hint: Call .WithAutoExecute(true) or use .WithReActAndTools()
```

---

## Additional Context

### User Journey

1. User reads documentation about ReAct mode ✅
2. User adds `WithReActMode(true)` ✅
3. User adds tools with `WithTool()` ✅
4. User runs code ✅
5. **Tools don't execute** ❌
6. **No error message** ❌
7. User spends hours debugging ❌
8. User discovers undocumented `WithAutoExecute()` ❌
9. Still doesn't work ❌
10. User discovers `WithParallelTools(false)` needed ❌
11. **Still doesn't work!** ❌

This is **poor developer experience** and could be solved with better defaults.

---

## Test Case

Here's a test case that should pass but currently fails:

```go
func TestReActWithTools(t *testing.T) {
    tool := agent.NewTool("add", "Add two numbers").
        AddParameter("a", "number", "First number", true).
        AddParameter("b", "number", "Second number", true)

    tool.Handler = func(args string) (string, error) {
        var params map[string]interface{}
        json.Unmarshal([]byte(args), &params)
        a := params["a"].(float64)
        b := params["b"].(float64)
        return fmt.Sprintf("%.0f", a+b), nil
    }

    // This should "just work"
    agent := agent.NewOpenAI("gpt-4o-mini", apiKey).
        WithDefaults().
        WithReActMode(true).
        WithTool(tool)

    result, err := agent.Execute(ctx, "What is 5 + 3?")

    assert.NoError(t, err)
    assert.True(t, result.Success)
    assert.Contains(t, result.Answer, "8")

    // Should have tool execution in steps
    hasToolExecution := false
    for _, step := range result.Steps {
        if step.Type == "observation" {
            hasToolExecution = true
            break
        }
    }
    assert.True(t, hasToolExecution, "Tool should have been executed")
}
```

**Current status:** ❌ FAILS - No tool execution

---

## Priority

**High Priority** - This affects core functionality and developer experience.

**Blocks:**
- Multi-agent workflows with tool verification
- Production use cases requiring calculation accuracy
- Any ReAct pattern with tools

**Impact:**
- 🔴 Core feature not working as expected
- 🔴 Poor developer experience
- 🔴 Confusing API
- 🔴 Silent failures

---

## Conclusion

The current API requires undocumented manual configuration for ReAct + Tools to work. This should be:

1. **Automatic** - `WithReActMode(true)` should enable required settings
2. **Documented** - If manual config needed, clearly document it
3. **Validated** - Detect misconfiguration and show helpful errors
4. **Consistent** - Examples should show working configurations

**Suggested immediate fixes:**

1. Make `WithReActMode(true)` auto-enable `WithAutoExecute(true)` and `WithParallelTools(false)`
2. Add validation that shows clear error messages
3. Update examples to show complete working configurations
4. Add documentation explaining why these settings are needed

This would dramatically improve developer experience and prevent confusion.

---

## Related Files

In our project, we've documented this issue in:
- `REACT_TOOL_ISSUE.md` - Detailed technical analysis
- `SOLUTION.md` - Our investigation of the problem
- `examples/with_math_tool.go` - Reproduction case
- `examples/full_workflow.go` - Production scenario

---

## Request

Could you please:

1. **Confirm if this is expected behavior** or a bug
2. **Provide the correct configuration** for ReAct + Tools
3. **Consider improving the API** to make this "just work"
4. **Update documentation** if manual configuration is required
5. **Add validation** to catch configuration errors early

Thank you for this excellent library! We're excited to use it, but this issue is currently blocking our production deployment.

---

**Reporter:** taipm
**Date:** 2025-11-11
**Library Version:** v0.7.2
**Issue Type:** Bug / API Design / Developer Experience
