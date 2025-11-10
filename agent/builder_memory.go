package agent

import (
	"github.com/taipm/go-deep-agent/agent/memory"
)

// Memory configuration methods for Builder
// This file contains all methods related to memory management,
// including hierarchical memory, episodic memory, and semantic memory.

func (b *Builder) WithMemory() *Builder {
	b.autoMemory = true
	return b
}

func (b *Builder) WithHierarchicalMemory(config memory.MemoryConfig) *Builder {
	b.memory = memory.NewWithConfig(config)
	b.memoryEnabled = true
	return b
}

func (b *Builder) DisableMemory() *Builder {
	b.memoryEnabled = false
	return b
}

func (b *Builder) GetMemory() *memory.Memory {
	return b.memory
}

func (b *Builder) WithEpisodicMemory(threshold float64) *Builder {
	if b.memory == nil {
		b.memory = memory.New()
	}
	config := b.memory.GetConfig()
	config.EpisodicEnabled = true
	config.EpisodicThreshold = threshold
	config.ImportanceScoring = true
	_ = b.memory.SetConfig(config) // Error only occurs if memory is nil, which we check above
	b.memoryEnabled = true
	return b
}

func (b *Builder) WithImportanceWeights(weights memory.ImportanceWeights) *Builder {
	if b.memory == nil {
		b.memory = memory.New()
	}
	config := b.memory.GetConfig()
	config.ImportanceWeights = weights
	config.ImportanceScoring = true
	_ = b.memory.SetConfig(config)
	b.memoryEnabled = true
	return b
}

func (b *Builder) WithWorkingMemorySize(size int) *Builder {
	if b.memory == nil {
		b.memory = memory.New()
	}
	config := b.memory.GetConfig()
	config.WorkingCapacity = size
	_ = b.memory.SetConfig(config)
	b.memoryEnabled = true
	return b
}

func (b *Builder) WithSemanticMemory() *Builder {
	if b.memory == nil {
		b.memory = memory.New()
	}
	config := b.memory.GetConfig()
	config.SemanticEnabled = true
	_ = b.memory.SetConfig(config)
	b.memoryEnabled = true
	return b
}
