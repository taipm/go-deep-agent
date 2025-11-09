#!/bin/bash

# Test script for chatbot_cli with Ollama
# This script tests if chatbot can connect to Ollama and get a response

echo "Testing chatbot_cli with Ollama (qwen2.5:1.5b)..."
echo ""

# Check if Ollama is running
if ! curl -s http://localhost:11434/api/tags > /dev/null 2>&1; then
    echo "❌ Error: Ollama is not running"
    echo "Please start Ollama with: ollama serve"
    exit 1
fi

echo "✅ Ollama is running"

# Check if qwen2.5:1.5b model is available
if ! ollama list | grep -q "qwen2.5:1.5b"; then
    echo "⚠️  Warning: qwen2.5:1.5b not found"
    echo "Pulling qwen2.5:1.5b..."
    ollama pull qwen2.5:1.5b
fi

echo "✅ qwen2.5:1.5b model is available"
echo ""
echo "Starting chatbot..."
echo "Commands to test:"
echo "  - Select option 4 (Ollama qwen2.5:1.5b)"
echo "  - Enable streaming: y"
echo "  - Enable memory: y"
echo "  - Type: hello"
echo "  - Type: /stats"
echo "  - Type: /exit"
echo ""
echo "Press Enter to continue..."
read

./chatbot_cli
