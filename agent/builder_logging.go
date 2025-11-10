package agent

// Logging configuration methods for Builder
// This file contains methods for configuring logging behavior.
// Note: injectLoggerToTools is in builder_tools.go

func (b *Builder) WithLogger(logger Logger) *Builder {
	b.logger = logger
	b.injectLoggerToTools() // Propagate logger to built-in tools
	return b
}

func (b *Builder) WithDebugLogging() *Builder {
	b.logger = NewStdLogger(LogLevelDebug)
	b.injectLoggerToTools() // Propagate logger to built-in tools
	return b
}

func (b *Builder) WithInfoLogging() *Builder {
	b.logger = NewStdLogger(LogLevelInfo)
	b.injectLoggerToTools() // Propagate logger to built-in tools
	return b
}

func (b *Builder) getLogger() Logger {
	if b.logger == nil {
		return &NoopLogger{}
	}
	return b.logger
}

// WithDebug enables enhanced debug mode with configurable logging.
// This is more comprehensive than WithDebugLogging() and includes:
//   - Request/response logging with secret redaction
//   - Error logging with full context
//   - Token usage tracking (verbose mode)
//   - Tool execution logging (verbose mode)
//   - Configurable log levels and truncation
//
// Example (basic debug mode):
//
//	builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
//	    WithDebug(agent.DefaultDebugConfig())
//
// Example (verbose debug mode):
//
//	builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
//	    WithDebug(agent.VerboseDebugConfig())
//
// Example (custom configuration):
//
//	config := agent.DebugConfig{
//	    Enabled:           true,
//	    Level:             agent.DebugLevelBasic,
//	    RedactSecrets:     true,
//	    LogRequests:       true,
//	    LogResponses:      true,
//	    LogErrors:         true,
//	    LogTokenUsage:     false,
//	    LogToolExecutions: false,
//	    MaxLogLength:      2000,
//	}
//	builder := agent.NewOpenAI("gpt-4o-mini", apiKey).
//	    WithDebug(config)
//
// Security Note: Always keep RedactSecrets=true in production to avoid
// logging API keys and sensitive data.
func (b *Builder) WithDebug(config DebugConfig) *Builder {
	b.debugConfig = config
	// Ensure we have a logger
	if b.logger == nil {
		b.logger = NewStdLogger(LogLevelDebug)
	}
	// Create debug logger instance
	b.debugLogger = newDebugLogger(config, b.logger)
	return b
}

// getDebugLogger returns the debug logger if debug mode is enabled.
func (b *Builder) getDebugLogger() *debugLogger {
	return b.debugLogger
}
