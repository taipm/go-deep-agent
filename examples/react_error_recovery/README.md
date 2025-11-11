# Error Recovery Example

Demonstrates ReAct error handling and recovery strategies.

## Features

- **Unreliable tool**: Fails 2 times, succeeds on 3rd attempt
- **Weird tool**: Returns non-JSON responses
- **Error tracking**: Count failures and show recovery
- **Retry logic**: LLM learns to retry failed operations

## Tools

### unreliable_service
Simulates network failures - fails attempts 1 & 2, succeeds on attempt 3.

### weird_tool  
Returns plain text instead of JSON to test observation parsing.

## Usage

```bash
export OPENAI_API_KEY="your-key"
go run main.go
```

## Expected Flow

**Task**: "Use unreliable_service to process 'data_backup'"

1. **ACTION**: unreliable_service → FAILED (attempt 1)
2. **THOUGHT**: Service unavailable, should retry
3. **ACTION**: unreliable_service → FAILED (attempt 2)
4. **THOUGHT**: Still failing, try again
5. **ACTION**: unreliable_service → SUCCESS (attempt 3)
6. **FINAL**: Operation completed after recovery

## Output

```
Error Recovery Example
======================

Task 1: Use unreliable_service to process 'data_backup'
---

Result:
  Success: true
  Iterations: 4
  Errors Encountered: 2

  Recovery Trace:
    1. unreliable_service → FAILED: service unavailable (attempt 1)
    2. unreliable_service → FAILED: service unavailable (attempt 2)
    3. unreliable_service → SUCCESS

  Answer: Operation completed successfully
```

## Key Learnings

- ReAct can recover from transient errors
- LLM learns retry strategies automatically
- Error context preserved in steps
- Non-strict mode allows parse recovery
