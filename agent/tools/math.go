// Package tools provides built-in tools for AI agents.
// This file implements MathTool - mathematical operations powered by professional libraries.
package tools

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/Knetic/govaluate"
	"github.com/taipm/go-deep-agent/agent"
	"gonum.org/v1/gonum/stat"
)

// NewMathTool creates a tool for mathematical operations.
// Powered by govaluate (expression evaluation) and gonum (statistics).
//
// Available operations:
//   - evaluate: Evaluate mathematical expressions with functions (sin, cos, sqrt, pow, log, etc.)
//   - statistics: Calculate statistical measures (mean, median, stdev, variance, min, max, sum)
//   - solve: Solve linear equations (quadratic support coming soon)
//   - convert: Convert between units (distance, weight, temperature, time)
//   - random: Generate random numbers (integer, float, choice from list)
//
// Example:
//
//	mathTool := tools.NewMathTool()
//	agent.NewOpenAI("gpt-4o", apiKey).
//	    WithTool(mathTool).
//	    WithAutoExecute().
//	    Ask(ctx, "Calculate: 2 * (3 + 4) + sqrt(16)")
func NewMathTool() *agent.Tool {
	tool := agent.NewTool("math", "Perform mathematical operations: expression evaluation, statistics, equation solving, unit conversion, random generation").
		AddParameter("operation", "string", "Operation: evaluate, statistics, solve, convert, random", true).
		AddParameter("expression", "string", "Math expression for evaluate (e.g., '2 * (3 + 4)', 'sin(3.14/2) + sqrt(16)')", false).
		AddParameter("stat_type", "string", "Statistics type: mean, median, stdev, variance, min, max, sum", false).
		AddParameter("equation", "string", "Equation to solve (e.g., 'x+5=10', 'x-3=7')", false).
		AddParameter("value", "number", "Value to convert", false).
		AddParameter("from_unit", "string", "Source unit (km, m, cm, kg, g, celsius, fahrenheit, hours, minutes, seconds)", false).
		AddParameter("to_unit", "string", "Target unit", false).
		AddParameter("random_type", "string", "Random type: integer, float, choice", false).
		AddParameter("min", "number", "Min value for random integer/float", false).
		AddParameter("max", "number", "Max value for random integer/float", false).
		WithHandler(mathHandler)

	// Manually add array parameters with proper items schema
	props := tool.Parameters["properties"].(map[string]interface{})
	props["numbers"] = agent.ArrayParam("Array of numbers for statistics", "number")
	props["choices"] = agent.ArrayParam("List of choices for random choice", "string")

	return tool
}

// mathHandler executes mathematical operations
func mathHandler(args string) (string, error) {
	var params struct {
		Operation  string    `json:"operation"`
		Expression string    `json:"expression"`
		Numbers    []float64 `json:"numbers"`
		StatType   string    `json:"stat_type"`
		Equation   string    `json:"equation"`
		Value      float64   `json:"value"`
		FromUnit   string    `json:"from_unit"`
		ToUnit     string    `json:"to_unit"`
		RandomType string    `json:"random_type"`
		Min        float64   `json:"min"`
		Max        float64   `json:"max"`
		Choices    []string  `json:"choices"`
	}

	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("%w: invalid JSON parameters", ErrInvalidInput)
	}

	switch params.Operation {
	case "evaluate":
		return evaluate(params.Expression)
	case "statistics":
		return statistics(params.Numbers, params.StatType)
	case "solve":
		return solve(params.Equation)
	case "convert":
		return convert(params.Value, params.FromUnit, params.ToUnit)
	case "random":
		return randomOp(params.RandomType, params.Min, params.Max, params.Choices)
	default:
		return "", fmt.Errorf("%w: unknown operation '%s'", ErrInvalidInput, params.Operation)
	}
}

// evaluate evaluates mathematical expressions using govaluate
func evaluate(expression string) (string, error) {
	ctx := getContext()

	logDebug(ctx, "Evaluating math expression", map[string]interface{}{
		"tool":       "math",
		"operation":  "evaluate",
		"expression": expression,
	})

	if expression == "" {
		logWarn(ctx, "Empty expression provided", map[string]interface{}{
			"tool":      "math",
			"operation": "evaluate",
		})
		return "", fmt.Errorf("%w: expression is required", ErrInvalidInput)
	}

	// Create expression with built-in functions
	expr, err := govaluate.NewEvaluableExpressionWithFunctions(expression, map[string]govaluate.ExpressionFunction{
		"sqrt": func(args ...interface{}) (interface{}, error) {
			return math.Sqrt(args[0].(float64)), nil
		},
		"pow": func(args ...interface{}) (interface{}, error) {
			return math.Pow(args[0].(float64), args[1].(float64)), nil
		},
		"sin": func(args ...interface{}) (interface{}, error) {
			return math.Sin(args[0].(float64)), nil
		},
		"cos": func(args ...interface{}) (interface{}, error) {
			return math.Cos(args[0].(float64)), nil
		},
		"tan": func(args ...interface{}) (interface{}, error) {
			return math.Tan(args[0].(float64)), nil
		},
		"log": func(args ...interface{}) (interface{}, error) {
			return math.Log10(args[0].(float64)), nil
		},
		"ln": func(args ...interface{}) (interface{}, error) {
			return math.Log(args[0].(float64)), nil
		},
		"abs": func(args ...interface{}) (interface{}, error) {
			return math.Abs(args[0].(float64)), nil
		},
		"ceil": func(args ...interface{}) (interface{}, error) {
			return math.Ceil(args[0].(float64)), nil
		},
		"floor": func(args ...interface{}) (interface{}, error) {
			return math.Floor(args[0].(float64)), nil
		},
		"round": func(args ...interface{}) (interface{}, error) {
			return math.Round(args[0].(float64)), nil
		},
	})

	if err != nil {
		logError(ctx, "Invalid math expression", map[string]interface{}{
			"tool":       "math",
			"operation":  "evaluate",
			"expression": expression,
			"error":      err.Error(),
		})
		return "", fmt.Errorf("%w: invalid expression: %v", ErrInvalidInput, err)
	}

	// Evaluate expression
	result, err := expr.Evaluate(nil)
	if err != nil {
		logError(ctx, "Math evaluation failed", map[string]interface{}{
			"tool":       "math",
			"operation":  "evaluate",
			"expression": expression,
			"error":      err.Error(),
		})
		return "", fmt.Errorf("%w: evaluation failed: %v", ErrOperationFailed, err)
	}

	// Convert result to float64
	var resultFloat float64
	switch v := result.(type) {
	case float64:
		resultFloat = v
	case int:
		resultFloat = float64(v)
	default:
		logError(ctx, "Unexpected result type from evaluation", map[string]interface{}{
			"tool":        "math",
			"operation":   "evaluate",
			"expression":  expression,
			"result_type": fmt.Sprintf("%T", result),
		})
		return "", fmt.Errorf("%w: unexpected result type", ErrOperationFailed)
	}

	logDebug(ctx, "Math expression evaluated successfully", map[string]interface{}{
		"tool":       "math",
		"operation":  "evaluate",
		"expression": expression,
		"result":     resultFloat,
	})

	return fmt.Sprintf("%.6f", resultFloat), nil
}

// statistics calculates statistical measures using gonum
func statistics(numbers []float64, statType string) (string, error) {
	if len(numbers) == 0 {
		return "", fmt.Errorf("%w: numbers array is required", ErrInvalidInput)
	}

	if statType == "" {
		return "", fmt.Errorf("%w: stat_type is required", ErrInvalidInput)
	}

	var result float64

	switch statType {
	case "mean":
		result = stat.Mean(numbers, nil)
	case "median":
		// Sort numbers for median calculation
		sorted := make([]float64, len(numbers))
		copy(sorted, numbers)
		result = median(sorted)
	case "stdev":
		mean := stat.Mean(numbers, nil)
		result = stat.StdDev(numbers, nil)
		_ = mean // mean is calculated internally by StdDev
	case "variance":
		result = stat.Variance(numbers, nil)
	case "min":
		result = min(numbers)
	case "max":
		result = max(numbers)
	case "sum":
		result = 0
		for _, n := range numbers {
			result += n
		}
	default:
		return "", fmt.Errorf("%w: unknown stat_type '%s'", ErrInvalidInput, statType)
	}

	return fmt.Sprintf("%.6f", result), nil
}

// solve solves linear and quadratic equations
func solve(equation string) (string, error) {
	if equation == "" {
		return "", fmt.Errorf("%w: equation is required", ErrInvalidInput)
	}

	// Parse equation format: "ax^2 + bx + c = 0" or "ax + b = 0"
	parts := strings.Split(equation, "=")
	if len(parts) != 2 {
		return "", fmt.Errorf("%w: equation must contain '='", ErrInvalidInput)
	}

	left := strings.TrimSpace(parts[0])
	right := strings.TrimSpace(parts[1])

	// Simple linear equation: x + 5 = 10
	if strings.Contains(left, "x") && !strings.Contains(left, "x^2") && !strings.Contains(left, "*") {
		return solveLinear(left, right)
	}

	// Quadratic equation: 2x^2 + 3x - 5 = 0
	if strings.Contains(left, "x^2") {
		return solveQuadratic(left, right)
	}

	return "", fmt.Errorf("%w: unsupported equation format", ErrInvalidInput)
}

// solveLinear solves linear equations (ax + b = c)
func solveLinear(left, right string) (string, error) {
	// Parse right side
	rightVal, err := strconv.ParseFloat(right, 64)
	if err != nil {
		return "", fmt.Errorf("%w: invalid right side value", ErrInvalidInput)
	}

	// Simple case: x + b = c or x - b = c
	left = strings.ReplaceAll(left, " ", "")

	if strings.HasPrefix(left, "x+") {
		b, _ := strconv.ParseFloat(left[2:], 64)
		x := rightVal - b
		return fmt.Sprintf("x = %.6f", x), nil
	}

	if strings.HasPrefix(left, "x-") {
		b, _ := strconv.ParseFloat(left[2:], 64)
		x := rightVal + b
		return fmt.Sprintf("x = %.6f", x), nil
	}

	if left == "x" {
		return fmt.Sprintf("x = %.6f", rightVal), nil
	}

	return "", fmt.Errorf("%w: unsupported linear equation format", ErrInvalidInput)
}

// solveQuadratic solves quadratic equations (ax^2 + bx + c = 0)
func solveQuadratic(left, right string) (string, error) {
	// For now, return a simple implementation message
	// Full quadratic solver would require more complex parsing
	return "", fmt.Errorf("%w: quadratic solver not yet implemented", ErrOperationFailed)
}

// convert converts between units
func convert(value float64, fromUnit, toUnit string) (string, error) {
	if fromUnit == "" || toUnit == "" {
		return "", fmt.Errorf("%w: from_unit and to_unit are required", ErrInvalidInput)
	}

	fromUnit = strings.ToLower(fromUnit)
	toUnit = strings.ToLower(toUnit)

	// Distance conversions
	distanceUnits := map[string]float64{
		"km": 1000.0,
		"m":  1.0,
		"cm": 0.01,
		"mm": 0.001,
	}

	// Weight conversions
	weightUnits := map[string]float64{
		"kg": 1000.0,
		"g":  1.0,
		"mg": 0.001,
	}

	// Temperature conversion
	if fromUnit == "celsius" && toUnit == "fahrenheit" {
		result := (value * 9 / 5) + 32
		return fmt.Sprintf("%.6f %s", result, toUnit), nil
	}
	if fromUnit == "fahrenheit" && toUnit == "celsius" {
		result := (value - 32) * 5 / 9
		return fmt.Sprintf("%.6f %s", result, toUnit), nil
	}

	// Time conversions
	timeUnits := map[string]float64{
		"hours":   3600.0,
		"minutes": 60.0,
		"seconds": 1.0,
	}

	// Try distance conversion
	if fromFactor, ok := distanceUnits[fromUnit]; ok {
		if toFactor, ok := distanceUnits[toUnit]; ok {
			result := (value * fromFactor) / toFactor
			return fmt.Sprintf("%.6f %s", result, toUnit), nil
		}
	}

	// Try weight conversion
	if fromFactor, ok := weightUnits[fromUnit]; ok {
		if toFactor, ok := weightUnits[toUnit]; ok {
			result := (value * fromFactor) / toFactor
			return fmt.Sprintf("%.6f %s", result, toUnit), nil
		}
	}

	// Try time conversion
	if fromFactor, ok := timeUnits[fromUnit]; ok {
		if toFactor, ok := timeUnits[toUnit]; ok {
			result := (value * fromFactor) / toFactor
			return fmt.Sprintf("%.6f %s", result, toUnit), nil
		}
	}

	return "", fmt.Errorf("%w: unsupported unit conversion from '%s' to '%s'", ErrInvalidInput, fromUnit, toUnit)
}

// randomOp generates random numbers
func randomOp(randomType string, minVal, maxVal float64, choices []string) (string, error) {
	if randomType == "" {
		return "", fmt.Errorf("%w: random_type is required", ErrInvalidInput)
	}

	// Seed random number generator
	rand.Seed(time.Now().UnixNano())

	switch randomType {
	case "integer":
		if minVal >= maxVal {
			return "", fmt.Errorf("%w: min must be less than max", ErrInvalidInput)
		}
		result := int(minVal) + rand.Intn(int(maxVal-minVal+1))
		return fmt.Sprintf("%d", result), nil

	case "float":
		if minVal >= maxVal {
			return "", fmt.Errorf("%w: min must be less than max", ErrInvalidInput)
		}
		result := minVal + rand.Float64()*(maxVal-minVal)
		return fmt.Sprintf("%.6f", result), nil

	case "choice":
		if len(choices) == 0 {
			return "", fmt.Errorf("%w: choices array is required", ErrInvalidInput)
		}
		idx := rand.Intn(len(choices))
		return choices[idx], nil

	default:
		return "", fmt.Errorf("%w: unknown random_type '%s'", ErrInvalidInput, randomType)
	}
}

// Helper functions

func median(numbers []float64) float64 {
	n := len(numbers)
	if n == 0 {
		return 0
	}

	// Simple bubble sort for small arrays
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			if numbers[i] > numbers[j] {
				numbers[i], numbers[j] = numbers[j], numbers[i]
			}
		}
	}

	if n%2 == 0 {
		return (numbers[n/2-1] + numbers[n/2]) / 2
	}
	return numbers[n/2]
}

func min(numbers []float64) float64 {
	if len(numbers) == 0 {
		return 0
	}
	minVal := numbers[0]
	for _, n := range numbers {
		if n < minVal {
			minVal = n
		}
	}
	return minVal
}

func max(numbers []float64) float64 {
	if len(numbers) == 0 {
		return 0
	}
	maxVal := numbers[0]
	for _, n := range numbers {
		if n > maxVal {
			maxVal = n
		}
	}
	return maxVal
}
