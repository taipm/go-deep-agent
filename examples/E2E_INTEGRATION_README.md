# End-to-End Memory Integration Test

This test demonstrates the complete hierarchical memory system integration with real OpenAI API calls.

## What It Tests

1. **Important Message Storage**: Messages with "remember" keywords are stored in episodic memory
2. **Casual Message Filtering**: Low-importance messages stay in working memory only
3. **Memory Recall**: Assistant can recall previously stored information
4. **Auto-Compression**: Working memory compresses when full
5. **Importance Scoring**: Verifies importance calculation is working correctly

## Prerequisites

- OpenAI API key with access to `gpt-4o-mini`
- Go 1.21 or later

## How to Run

```bash
# Set your OpenAI API key
export OPENAI_API_KEY=sk-xxxxx

# Run the test
cd /Users/taipm/GitHub/go-deep-agent
go run examples/e2e_integration.go
```

## Expected Output

The test will:
1. Send several messages to OpenAI
2. Track memory statistics after each interaction
3. Verify that important messages are stored in episodic memory
4. Verify that casual messages are NOT stored in episodic
5. Print a final verification summary

## Success Criteria

All 5 verification tests should pass:
- ✅ Important messages stored in episodic memory
- ✅ Casual messages NOT stored in episodic
- ✅ All messages processed through memory system
- ✅ Working memory stays within capacity
- ✅ Average importance meets threshold

## Example Run

```
=== End-to-End Memory Integration Test with OpenAI ===

Test 1: User explicitly asks to remember something
User: Remember that my birthday is May 5th
Assistant: I'll remember that your birthday is on May 5th!

📊 Stats after Test 1:
  Working memory: 2 messages
  Episodic memory: 1 messages
  Average importance: 1.00

Test 2: Casual conversation (should not store in episodic)
User: What's the weather like?
Assistant: I don't have access to real-time weather information...

📊 Stats after Test 2:
  Working memory: 4 messages
  Episodic memory: 1 messages (should be same as Test 1)
  Average importance: 1.00

...

=== Results: 5/5 tests passed ===
🎉 All end-to-end integration tests PASSED!
```

## Configuration

The test uses:
- Working memory capacity: 5 messages
- Episodic threshold: 0.7
- Model: gpt-4o-mini

You can modify these in the code if needed.

## Notes

- This test makes actual API calls and will consume OpenAI credits
- Each test run makes ~7-8 API requests
- Estimated cost: < $0.01 per run with gpt-4o-mini
- The test includes 1-second delays between calls to allow memory processing
