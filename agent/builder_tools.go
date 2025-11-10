package agent

import (
	"time"

	"github.com/openai/openai-go/v3"
)

// Tool configuration methods for Builder
// This file contains all methods related to tool management,
// parallel execution, and tool callbacks.

func (b *Builder) OnToolCall(callback func(openai.FinishedChatCompletionToolCall)) *Builder {
	b.onToolCall = callback
	return b
}

func (b *Builder) WithTool(tool *Tool) *Builder {
	b.tools = append(b.tools, tool)
	return b
}

func (b *Builder) WithTools(tools ...*Tool) *Builder {
	b.tools = append(b.tools, tools...)
	return b
}

func (b *Builder) WithAutoExecute(enable bool) *Builder {
	b.autoExecute = enable
	if b.maxToolRounds == 0 {
		b.maxToolRounds = 5 // Default max rounds
	}
	return b
}

func (b *Builder) WithMaxToolRounds(max int) *Builder {
	b.maxToolRounds = max
	return b
}

func (b *Builder) WithParallelTools(enable bool) *Builder {
	b.enableParallel = enable
	if b.maxWorkers == 0 {
		b.maxWorkers = 10 // Default worker pool size
	}
	if b.toolTimeout == 0 {
		b.toolTimeout = 30 * time.Second // Default timeout
	}
	return b
}

func (b *Builder) WithMaxWorkers(max int) *Builder {
	b.maxWorkers = max
	return b
}

func (b *Builder) WithToolTimeout(timeout time.Duration) *Builder {
	b.toolTimeout = timeout
	return b
}

func (b *Builder) injectLoggerToTools() {
	// TODO: Fix go:linkname relocation issue in tests
	// Temporarily disabled to allow tests to run
	/*
		logger := b.getLogger()

		// Inject callback function to tools package via go:linkname
		toolsSetLogFunc(func(level, msg string, fields map[string]interface{}) {
			ctx := context.Background() // Tools don't have context, use background

			// Convert map[string]interface{} to []Field
			logFields := make([]Field, 0, len(fields))
			for k, v := range fields {
				logFields = append(logFields, F(k, v))
			}

			// Route to appropriate log level
			switch level {
			case "DEBUG":
				logger.Debug(ctx, msg, logFields...)
			case "INFO":
				logger.Info(ctx, msg, logFields...)
			case "WARN":
				logger.Warn(ctx, msg, logFields...)
			case "ERROR":
				logger.Error(ctx, msg, logFields...)
			}
		})
	*/
}
