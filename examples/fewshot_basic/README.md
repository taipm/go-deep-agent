# Few-Shot Learning Example

This example demonstrates how to use few-shot learning with go-deep-agent.

## Features Demonstrated

- Adding few-shot examples inline using `AddFewShotExample()`
- French translation with example pairs
- Improved consistency through few-shot prompting

## Running the Example

```bash
export OPENAI_API_KEY=your_api_key_here
go run main.go
```

## Expected Output

```
=== Few-Shot Translation Example ===
Translation: Bonjour (or similar French greeting)
```

## How It Works

The agent is configured with 3 few-shot examples:
1. "Translate: Hello" → "Bonjour"
2. "Translate: Goodbye" → "Au revoir"  
3. "Translate: Thank you" → "Merci"

These examples are automatically injected into the prompt before your query, helping the LLM understand the expected format and style of translations.

## Related Documentation

- See `../../personas/translator_fewshot.yaml` for YAML-based few-shot configuration
- See `agent/fewshot.go` for implementation details
- See `agent/fewshot_test.go` for more usage examples
