#!/bin/bash

# Test memory functionality
# Simulate a conversation to verify memory works

echo "Testing Math Teacher Memory..."
echo ""

# Create test input
cat > /tmp/math_teacher_test_input.txt << 'EOF'
Tên con là Lan
Bạn nhớ tên con chưa?
exit
EOF

echo "Input test:"
cat /tmp/math_teacher_test_input.txt
echo ""
echo "Running test..."
echo ""

# Run with test input
go run . interactive < /tmp/math_teacher_test_input.txt

# Cleanup
rm /tmp/math_teacher_test_input.txt
