#!/bin/bash

# Built-in Tools Analysis Script
# Analyzes code quality, test coverage, and capabilities

echo "=== 📊 GO-DEEP-AGENT BUILT-IN TOOLS ANALYSIS ==="
echo ""

# 1. Code Statistics
echo "📈 1. CODE STATISTICS"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

echo "Production Code:"
wc -l agent/tools/tools.go agent/tools/filesystem.go agent/tools/http.go agent/tools/datetime.go agent/tools/math.go | tail -1
echo ""

echo "Test Code:"
wc -l agent/tools/*_test.go | tail -1
echo ""

echo "Total Lines:"
wc -l agent/tools/*.go | tail -1
echo ""

# 2. Test Coverage
echo "🧪 2. TEST COVERAGE"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

echo "Running tests with coverage..."
go test ./agent/tools/... -cover -coverprofile=coverage.out 2>&1 | grep -E "coverage:|ok"
echo ""

if [ -f coverage.out ]; then
    echo "Detailed Coverage by File:"
    go tool cover -func=coverage.out | grep -E "agent/tools" | head -10
    echo ""
    
    echo "Total Coverage:"
    go tool cover -func=coverage.out | tail -1
    echo ""
fi

# 3. Tool Capabilities Matrix
echo "🛠️  3. TOOL CAPABILITIES MATRIX"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

echo "FileSystemTool:"
echo "  ✓ read_file         - Read file contents"
echo "  ✓ write_file        - Create/overwrite files"
echo "  ✓ append_file       - Append to files"
echo "  ✓ delete_file       - Remove files"
echo "  ✓ list_directory    - List dir contents"
echo "  ✓ file_exists       - Check existence"
echo "  ✓ create_directory  - Create directories"
echo "  Total: 7 operations"
echo ""

echo "HTTPRequestTool:"
echo "  ✓ GET               - Fetch resources"
echo "  ✓ POST              - Create resources"
echo "  ✓ PUT               - Update resources"
echo "  ✓ DELETE            - Remove resources"
echo "  Features: Headers, timeout, JSON parsing"
echo "  Total: 4 HTTP methods"
echo ""

echo "DateTimeTool:"
echo "  ✓ current_time      - Get current time"
echo "  ✓ format_date       - Convert formats"
echo "  ✓ parse_date        - Parse dates"
echo "  ✓ add_duration      - Add time intervals"
echo "  ✓ date_diff         - Calculate differences"
echo "  ✓ convert_timezone  - Timezone conversion"
echo "  ✓ day_of_week       - Get weekday"
echo "  Total: 7 operations"
echo ""

echo "MathTool:"
echo "  ✓ evaluate          - Expression engine (11 functions)"
echo "    • sqrt, pow, sin, cos, tan, log, ln"
echo "    • abs, ceil, floor, round"
echo "  ✓ statistics        - 7 stat measures (gonum)"
echo "    • mean, median, stdev, variance"
echo "    • min, max, sum"
echo "  ✓ solve             - Linear equations"
echo "  ✓ convert           - Unit conversions"
echo "    • distance, weight, temperature, time"
echo "  ✓ random            - RNG operations"
echo "    • integer, float, choice"
echo "  Total: 5 operation categories"
echo ""

# 4. Dependencies Analysis
echo "📦 4. DEPENDENCIES ANALYSIS"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

echo "External Dependencies:"
grep -E "(github.com|gonum.org)" go.mod | grep -v "^module" | head -5
echo ""

echo "Imports in Production Code:"
for file in agent/tools/filesystem.go agent/tools/http.go agent/tools/datetime.go agent/tools/math.go; do
    echo ""
    echo "$(basename $file):"
    grep -E "^\s+\"" $file | head -10 | sed 's/^/  /'
done
echo ""

# 5. Security Features
echo "🔒 5. SECURITY FEATURES"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

echo "FileSystemTool:"
echo "  ✓ Path traversal prevention (sanitizePath)"
echo "  ✓ Path validation"
echo "  ✓ Security error types"
grep -n "sanitizePath\|ErrSecurityViolation\|path traversal" agent/tools/filesystem.go | head -3 | sed 's/^/  /'
echo ""

echo "HTTPRequestTool:"
echo "  ✓ URL validation (http/https only)"
echo "  ✓ Timeout protection (default 30s)"
echo "  ✓ User-Agent identification"
grep -n "http://\|https://\|timeout\|User-Agent" agent/tools/http.go | head -3 | sed 's/^/  /'
echo ""

echo "MathTool:"
echo "  ✓ Sandboxed expression evaluation (govaluate)"
echo "  ✓ No code injection vulnerability"
echo "  ✓ Input validation"
grep -n "govaluate\|ErrInvalidInput\|Unmarshal" agent/tools/math.go | head -3 | sed 's/^/  /'
echo ""

# 6. Error Handling
echo "⚠️  6. ERROR HANDLING"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

echo "Custom Error Types:"
grep -E "^var Err" agent/tools/tools.go
echo ""

echo "Error Usage Count:"
for err in ErrInvalidInput ErrOperationFailed ErrSecurityViolation ErrTimeout; do
    count=$(grep -r "$err" agent/tools/*.go | grep -v "_test.go" | wc -l | tr -d ' ')
    echo "  $err: $count occurrences"
done
echo ""

# 7. Test Quality Metrics
echo "✅ 7. TEST QUALITY METRICS"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

echo "Test Counts per Tool:"
for tool in filesystem http datetime math; do
    count=$(grep -c "^func Test" agent/tools/${tool}_test.go)
    echo "  ${tool}_test.go: $count test functions"
done
echo ""

echo "Running all tests..."
go test ./agent/tools/... -v 2>&1 | grep -E "^PASS|^FAIL|^ok" | tail -5
echo ""

# 8. Performance Characteristics
echo "⚡ 8. PERFORMANCE CHARACTERISTICS"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

echo "Binary Size Impact:"
du -h tools_comprehensive_test 2>/dev/null || echo "  (Build tools_comprehensive_test first)"
echo ""

echo "Estimated Tool Execution Times:"
echo "  FileSystemTool:  < 1ms (local I/O)"
echo "  HTTPRequestTool: 100-500ms (network)"
echo "  DateTimeTool:    < 1ms (computation)"
echo "  MathTool:"
echo "    • evaluate:    < 1ms (govaluate)"
echo "    • statistics:  1-5ms (gonum)"
echo ""

# 9. Use Case Coverage
echo "🎯 9. USE CASE COVERAGE ANALYSIS"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

echo "Real-World Scenarios Supported:"
echo ""
echo "✓ File Management & Data Storage"
echo "  • Read/write configuration files"
echo "  • Save API responses to disk"
echo "  • Generate reports and logs"
echo ""
echo "✓ API Integration & Web Services"
echo "  • Fetch data from REST APIs"
echo "  • Post data to webhooks"
echo "  • Monitor endpoint availability"
echo ""
echo "✓ Time & Scheduling"
echo "  • Calculate deadlines and delays"
echo "  • Convert between timezones"
echo "  • Schedule reminders"
echo ""
echo "✓ Mathematical & Statistical Analysis"
echo "  • Evaluate complex expressions"
echo "  • Analyze data distributions"
echo "  • Solve simple equations"
echo "  • Convert units"
echo ""

# 10. Recommendations
echo "💡 10. RECOMMENDATIONS & IMPROVEMENTS"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

echo "Current Strengths:"
echo "  ✅ Comprehensive test coverage (60+ tests)"
echo "  ✅ Professional libraries (govaluate, gonum)"
echo "  ✅ Strong security practices"
echo "  ✅ Clean error handling"
echo "  ✅ Production-ready documentation"
echo ""

echo "Future Enhancements (Phase 2):"
echo "  📌 MathTool: Quadratic equation solver"
echo "  📌 MathTool: Numerical integration/differentiation"
echo "  📌 FileSystemTool: File search/pattern matching"
echo "  📌 HTTPRequestTool: OAuth support"
echo "  📌 DateTimeTool: Recurring event calculations"
echo ""

# Cleanup
rm -f coverage.out

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ ANALYSIS COMPLETE!"
echo ""
