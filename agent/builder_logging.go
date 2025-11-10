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
