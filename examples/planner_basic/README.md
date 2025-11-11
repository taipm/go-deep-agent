# Basic Planning Example

This example demonstrates the planning capabilities of go-deep-agent.

## Overview

The agent will:
1. Decompose a complex goal into subtasks
2. Execute tasks with dependency management  
3. Track progress and collect metrics
4. Return structured results

## Running

```bash
export OPENAI_API_KEY="your-api-key"
cd examples/planner_basic
go run main.go
```

## Expected Output

The agent will:
- Break down the research goal into tasks
- Execute each task sequentially
- Show the plan structure and execution metrics
- Display the final synthesized result

## Key Features Demonstrated

- **Automatic decomposition**: Goal → Task tree
- **Dependency handling**: Sequential execution
- **Metrics collection**: Duration, success rate
- **Structured output**: PlanResult with details
