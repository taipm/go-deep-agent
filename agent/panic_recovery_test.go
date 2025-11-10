package agent

import (
	"errors"
	"strings"
	"testing"
)

func TestPanicError_Error(t *testing.T) {
	panicErr := &PanicError{
		Value:      "something went wrong",
		StackTrace: "stack trace here",
	}

	msg := panicErr.Error()
	if !strings.Contains(msg, "panic recovered") {
		t.Errorf("Expected error message to contain 'panic recovered', got: %s", msg)
	}
	if !strings.Contains(msg, "something went wrong") {
		t.Errorf("Expected error message to contain panic value, got: %s", msg)
	}
}

func TestPanicError_Unwrap(t *testing.T) {
	panicErr := &PanicError{
		Value:      "test",
		StackTrace: "trace",
	}

	if panicErr.Unwrap() != nil {
		t.Error("Expected Unwrap to return nil")
	}
}

func TestRecoverPanic_Success(t *testing.T) {
	var err error
	func() {
		defer recoverPanic(&err, "test function")
		// No panic - should complete normally
	}()

	if err != nil {
		t.Errorf("Expected no error when no panic, got: %v", err)
	}
}

func TestRecoverPanic_StringPanic(t *testing.T) {
	var err error
	func() {
		defer recoverPanic(&err, "test function")
		panic("test panic")
	}()

	if err == nil {
		t.Fatal("Expected error from recovered panic")
	}

	panicErr, ok := err.(*PanicError)
	if !ok {
		t.Fatalf("Expected PanicError, got: %T", err)
	}

	if panicErr.Value != "test panic" {
		t.Errorf("Expected panic value 'test panic', got: %v", panicErr.Value)
	}

	if panicErr.StackTrace == "" {
		t.Error("Expected stack trace to be captured")
	}

	if !strings.Contains(panicErr.StackTrace, "panic_recovery_test.go") {
		t.Error("Expected stack trace to contain source file name")
	}
}

func TestRecoverPanic_ErrorPanic(t *testing.T) {
	var err error
	func() {
		defer recoverPanic(&err, "test function")
		panic(errors.New("test error"))
	}()

	if err == nil {
		t.Fatal("Expected error from recovered panic")
	}

	panicErr, ok := err.(*PanicError)
	if !ok {
		t.Fatalf("Expected PanicError, got: %T", err)
	}

	// Panic value should be the original error
	panicValue, ok := panicErr.Value.(error)
	if !ok {
		t.Fatalf("Expected panic value to be an error, got: %T", panicErr.Value)
	}

	if panicValue.Error() != "test error" {
		t.Errorf("Expected panic value 'test error', got: %v", panicValue)
	}
}

func TestRecoverPanic_IntPanic(t *testing.T) {
	var err error
	func() {
		defer recoverPanic(&err, "test function")
		panic(42)
	}()

	if err == nil {
		t.Fatal("Expected error from recovered panic")
	}

	panicErr, ok := err.(*PanicError)
	if !ok {
		t.Fatalf("Expected PanicError, got: %T", err)
	}

	if panicErr.Value != 42 {
		t.Errorf("Expected panic value 42, got: %v", panicErr.Value)
	}
}

func TestRecoverPanicWithLogger(t *testing.T) {
	mock := &mockLogger{}
	var err error

	func() {
		defer recoverPanicWithLogger(&err, "test context", mock)
		panic("logged panic")
	}()

	if err == nil {
		t.Fatal("Expected error from recovered panic")
	}

	// Check that panic was logged
	if len(mock.errorMsgs) != 1 {
		t.Fatalf("Expected 1 error log, got %d", len(mock.errorMsgs))
	}

	logMsg := mock.errorMsgs[0]
	if !strings.Contains(logMsg, "PANIC RECOVERED") {
		t.Error("Expected log to contain 'PANIC RECOVERED'")
	}
	if !strings.Contains(logMsg, "test context") {
		t.Error("Expected log to contain context")
	}
	if !strings.Contains(logMsg, "logged panic") {
		t.Error("Expected log to contain panic value")
	}
	if !strings.Contains(logMsg, "Stack trace:") {
		t.Error("Expected log to contain stack trace")
	}
}

func TestRecoverPanicWithLogger_NoLogger(t *testing.T) {
	var err error

	// Should not crash even with nil logger
	func() {
		defer recoverPanicWithLogger(&err, "test context", nil)
		panic("no logger panic")
	}()

	if err == nil {
		t.Fatal("Expected error from recovered panic")
	}
}

func TestSafeExecute_Success(t *testing.T) {
	result, err := safeExecute("test", func() (string, error) {
		return "success", nil
	})

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if result != "success" {
		t.Errorf("Expected result 'success', got: %s", result)
	}
}

func TestSafeExecute_Error(t *testing.T) {
	result, err := safeExecute("test", func() (string, error) {
		return "", errors.New("function error")
	})

	if err == nil {
		t.Fatal("Expected error")
	}
	if err.Error() != "function error" {
		t.Errorf("Expected 'function error', got: %v", err)
	}
	if result != "" {
		t.Errorf("Expected empty result, got: %s", result)
	}
}

func TestSafeExecute_Panic(t *testing.T) {
	result, err := safeExecute("test", func() (string, error) {
		panic("function panic")
	})

	if err == nil {
		t.Fatal("Expected error from panic")
	}

	panicErr, ok := err.(*PanicError)
	if !ok {
		t.Fatalf("Expected PanicError, got: %T", err)
	}

	if panicErr.Value != "function panic" {
		t.Errorf("Expected panic value 'function panic', got: %v", panicErr.Value)
	}

	if result != "" {
		t.Errorf("Expected empty result, got: %s", result)
	}
}

func TestSafeExecuteWithLogger(t *testing.T) {
	mock := &mockLogger{}

	result, err := safeExecuteWithLogger("test context", mock, func() (string, error) {
		panic("logged execute panic")
	})

	if err == nil {
		t.Fatal("Expected error from panic")
	}

	// Check logging
	if len(mock.errorMsgs) != 1 {
		t.Fatalf("Expected 1 error log, got %d", len(mock.errorMsgs))
	}

	if !strings.Contains(mock.errorMsgs[0], "test context") {
		t.Error("Expected log to contain context")
	}

	if result != "" {
		t.Errorf("Expected empty result, got: %s", result)
	}
}

func TestSafeExecuteVoid_Success(t *testing.T) {
	err := safeExecuteVoid("test", func() error {
		return nil
	})

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestSafeExecuteVoid_Error(t *testing.T) {
	err := safeExecuteVoid("test", func() error {
		return errors.New("void error")
	})

	if err == nil {
		t.Fatal("Expected error")
	}
	if err.Error() != "void error" {
		t.Errorf("Expected 'void error', got: %v", err)
	}
}

func TestSafeExecuteVoid_Panic(t *testing.T) {
	err := safeExecuteVoid("test", func() error {
		panic("void panic")
		return nil
	})

	if err == nil {
		t.Fatal("Expected error from panic")
	}

	panicErr, ok := err.(*PanicError)
	if !ok {
		t.Fatalf("Expected PanicError, got: %T", err)
	}

	if panicErr.Value != "void panic" {
		t.Errorf("Expected panic value 'void panic', got: %v", panicErr.Value)
	}
}

func TestSafeExecuteVoidWithLogger(t *testing.T) {
	mock := &mockLogger{}

	err := safeExecuteVoidWithLogger("void context", mock, func() error {
		panic("void logged panic")
	})

	if err == nil {
		t.Fatal("Expected error from panic")
	}

	// Check logging
	if len(mock.errorMsgs) != 1 {
		t.Fatalf("Expected 1 error log, got %d", len(mock.errorMsgs))
	}

	if !strings.Contains(mock.errorMsgs[0], "void context") {
		t.Error("Expected log to contain context")
	}
}

func TestIsPanicError(t *testing.T) {
	panicErr := &PanicError{Value: "test", StackTrace: "trace"}
	normalErr := errors.New("normal error")

	if !IsPanicError(panicErr) {
		t.Error("Expected IsPanicError to return true for PanicError")
	}

	if IsPanicError(normalErr) {
		t.Error("Expected IsPanicError to return false for normal error")
	}

	if IsPanicError(nil) {
		t.Error("Expected IsPanicError to return false for nil")
	}
}

func TestGetPanicValue(t *testing.T) {
	panicErr := &PanicError{Value: "test value", StackTrace: "trace"}
	normalErr := errors.New("normal error")

	value := GetPanicValue(panicErr)
	if value != "test value" {
		t.Errorf("Expected 'test value', got: %v", value)
	}

	value = GetPanicValue(normalErr)
	if value != nil {
		t.Errorf("Expected nil for normal error, got: %v", value)
	}

	value = GetPanicValue(nil)
	if value != nil {
		t.Errorf("Expected nil for nil error, got: %v", value)
	}
}

func TestGetStackTrace(t *testing.T) {
	panicErr := &PanicError{Value: "test", StackTrace: "test stack trace"}
	normalErr := errors.New("normal error")

	trace := GetStackTrace(panicErr)
	if trace != "test stack trace" {
		t.Errorf("Expected 'test stack trace', got: %s", trace)
	}

	trace = GetStackTrace(normalErr)
	if trace != "" {
		t.Errorf("Expected empty string for normal error, got: %s", trace)
	}

	trace = GetStackTrace(nil)
	if trace != "" {
		t.Errorf("Expected empty string for nil error, got: %s", trace)
	}
}

func TestPanicRecovery_RealWorldScenario(t *testing.T) {
	mock := &mockLogger{}

	// Simulate a tool handler that panics
	toolHandler := func(args string) (string, error) {
		// Simulate a panic (e.g., nil pointer, array out of bounds)
		var arr []int
		_ = arr[10] // This will panic with index out of range
		return "never reached", nil
	}

	// Wrap with safe execution
	result, err := safeExecuteWithLogger("tool: calculator", mock, func() (string, error) {
		return toolHandler("test args")
	})

	// Should recover from panic
	if err == nil {
		t.Fatal("Expected panic to be recovered as error")
	}

	// Should be a PanicError
	if !IsPanicError(err) {
		t.Errorf("Expected PanicError, got: %T", err)
	}

	// Should log the panic
	if len(mock.errorMsgs) != 1 {
		t.Fatalf("Expected panic to be logged, got %d logs", len(mock.errorMsgs))
	}

	// Result should be empty
	if result != "" {
		t.Errorf("Expected empty result after panic, got: %s", result)
	}

	// Log should contain useful info
	logMsg := mock.errorMsgs[0]
	if !strings.Contains(logMsg, "tool: calculator") {
		t.Error("Expected log to contain tool context")
	}
	if !strings.Contains(logMsg, "PANIC RECOVERED") {
		t.Error("Expected log to indicate panic recovery")
	}
}

func TestPanicRecovery_NestedPanics(t *testing.T) {
	var err error

	func() {
		defer recoverPanic(&err, "outer")
		func() {
			defer recoverPanic(&err, "inner")
			panic("inner panic")
		}()
		// Inner panic was recovered, outer should not see it
		if err == nil {
			t.Error("Inner panic should have been set")
		}
	}()

	// Only inner panic should be captured
	panicErr, ok := err.(*PanicError)
	if !ok {
		t.Fatalf("Expected PanicError, got: %T", err)
	}

	if panicErr.Value != "inner panic" {
		t.Errorf("Expected 'inner panic', got: %v", panicErr.Value)
	}
}

func TestPanicRecovery_NilPointerDereference(t *testing.T) {
	var str *string
	var err error

	func() {
		defer recoverPanic(&err, "nil pointer test")
		_ = *str // This will panic with nil pointer dereference
	}()

	if err == nil {
		t.Fatal("Expected panic from nil pointer dereference")
	}

	if !IsPanicError(err) {
		t.Errorf("Expected PanicError, got: %T", err)
	}

	// The panic value should be runtime.errorString
	stackTrace := GetStackTrace(err)
	if stackTrace == "" {
		t.Error("Expected stack trace to be captured")
	}
}
