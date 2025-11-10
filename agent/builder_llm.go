package agent

// LLM parameter configuration methods for Builder
// This file contains all methods related to LLM parameters
// such as temperature, top_p, max_tokens, penalties, etc.

func (b *Builder) WithTemperature(temperature float64) *Builder {
	b.temperature = &temperature
	return b
}

func (b *Builder) WithTopP(topP float64) *Builder {
	b.topP = &topP
	return b
}

func (b *Builder) WithMaxTokens(maxTokens int64) *Builder {
	b.maxTokens = &maxTokens
	return b
}

func (b *Builder) WithPresencePenalty(penalty float64) *Builder {
	b.presencePenalty = &penalty
	return b
}

func (b *Builder) WithFrequencyPenalty(penalty float64) *Builder {
	b.frequencyPenalty = &penalty
	return b
}

func (b *Builder) WithSeed(seed int64) *Builder {
	b.seed = &seed
	return b
}

func (b *Builder) WithLogprobs(enable bool) *Builder {
	b.logprobs = &enable
	return b
}

func (b *Builder) WithTopLogprobs(n int64) *Builder {
	b.topLogprobs = &n
	return b
}

func (b *Builder) WithMultipleChoices(n int64) *Builder {
	b.n = &n
	return b
}
