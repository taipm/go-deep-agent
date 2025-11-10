package agent

// Callback configuration methods for Builder
// This file contains methods for setting up callbacks for various events
// during LLM execution (streaming, refusals).
// Note: OnToolCall is in builder_tools.go

func (b *Builder) OnStream(callback func(string)) *Builder {
	b.onStream = callback
	return b
}

func (b *Builder) OnRefusal(callback func(string)) *Builder {
	b.onRefusal = callback
	return b
}
