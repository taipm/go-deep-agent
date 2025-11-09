#!/bin/bash

# Test conversation memory với qwen3:1.7b

echo "🧪 Testing Conversation Memory with qwen3:1.7b"
echo "=============================================="
echo ""
echo "Bạn sẽ test như sau:"
echo "1. Chọn option 4 (qwen3:1.7b)"
echo "2. Enable streaming: y"
echo "3. Enable memory: y"
echo ""
echo "📝 Test scenario 1: Simple name memory"
echo "   You: My name is John"
echo "   AI: [should greet John]"
echo "   You: /history    <- Check history"
echo "   You: What is my name?"
echo "   AI: [should say 'John']"
echo ""
echo "📝 Test scenario 2: Vietnamese context"
echo "   You: Tôi là Phan Minh Tài"
echo "   AI: [should greet]"
echo "   You: /history    <- Verify 2 messages saved"
echo "   You: Tôi tên gì?"
echo "   AI: [should remember 'Phan Minh Tài']"
echo ""
echo "📝 Test scenario 3: Number memory"
echo "   You: Remember this: 42"
echo "   AI: [confirms]"
echo "   You: What number did I tell you?"
echo "   AI: [should say '42']"
echo ""
echo "🔍 If memory doesn't work:"
echo "   - Use /history to see if messages are saved"
echo "   - Check if history shows both User and AI messages"
echo "   - Model may have context window limitations"
echo ""
echo "Press Enter to start chatbot..."
read

cd /Users/taipm/GitHub/go-deep-agent/examples
./chatbot_cli
