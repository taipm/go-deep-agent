package memory

import (
	"context"
	"time"
)

// Message represents a conversation message with metadata
type Message struct {
	Role      string                 // "system", "user", "assistant"
	Content   string                 // Message content
	Timestamp time.Time              // When message was created
	Metadata  map[string]interface{} // Additional metadata (importance, tags, etc.)
}

// MemorySystem is the main interface for hierarchical memory management
// It coordinates between Working, Episodic, and Semantic memory tiers
type MemorySystem interface {
	// Add a message to the memory system
	// The system automatically determines which tier(s) to store it in
	Add(ctx context.Context, msg Message) error

	// Recall retrieves relevant messages based on query and options
	// Combines results from all memory tiers intelligently
	Recall(ctx context.Context, query string, opts RecallOptions) ([]Message, error)

	// Compress triggers memory compression/summarization
	// Moves old working memory to episodic, summarizes low-importance messages
	Compress(ctx context.Context) error

	// Clear removes all messages from all tiers
	Clear(ctx context.Context) error

	// Stats returns current memory statistics
	Stats(ctx context.Context) MemoryStats

	// SetConfig updates memory system configuration
	SetConfig(config MemoryConfig) error

	// GetConfig returns current configuration
	GetConfig() MemoryConfig
}

// WorkingMemory represents short-term, hot memory (recent context)
// Equivalent to human "working memory" - limited capacity, fast access
type WorkingMemory interface {
	// Add a message to working memory
	Add(ctx context.Context, msg Message) error

	// Recent returns the N most recent messages
	Recent(ctx context.Context, n int) ([]Message, error)

	// All returns all messages in working memory
	All(ctx context.Context) ([]Message, error)

	// Clear removes all messages
	Clear(ctx context.Context) error

	// Compress summarizes old messages when capacity is exceeded
	// Returns summarized message and list of messages that were compressed
	Compress(ctx context.Context) (summary Message, compressed []Message, error error)

	// Size returns current number of messages in working memory
	Size() int

	// Capacity returns maximum capacity
	Capacity() int
}

// EpisodicMemory represents event-based, searchable memory
// Stores important past conversations with vector-based retrieval
type EpisodicMemory interface {
	// Store a message in episodic memory with importance score
	Store(ctx context.Context, msg Message, importance float64) error

	// StoreBatch stores multiple messages efficiently
	StoreBatch(ctx context.Context, messages []Message, importances []float64) error

	// Retrieve finds similar messages based on semantic similarity
	Retrieve(ctx context.Context, query string, topK int) ([]Message, error)

	// RetrieveByTime finds messages within a time range
	RetrieveByTime(ctx context.Context, start, end time.Time, limit int) ([]Message, error)

	// RetrieveByImportance finds messages above importance threshold
	RetrieveByImportance(ctx context.Context, minImportance float64, limit int) ([]Message, error)

	// Search combines multiple filters (similarity + time + importance)
	Search(ctx context.Context, filter SearchFilter) ([]Message, error)

	// Clear removes all episodic memories
	Clear(ctx context.Context) error

	// Size returns total number of stored messages
	Size() int
}

// SemanticMemory represents long-term knowledge and facts
// Stores learned information, preferences, and domain knowledge
type SemanticMemory interface {
	// StoreFact stores a fact or knowledge item
	StoreFact(ctx context.Context, fact Fact) error

	// QueryKnowledge retrieves relevant facts based on query
	QueryKnowledge(ctx context.Context, query string, limit int) ([]Fact, error)

	// UpdateFact updates an existing fact
	UpdateFact(ctx context.Context, factID string, fact Fact) error

	// DeleteFact removes a fact
	DeleteFact(ctx context.Context, factID string) error

	// ListFacts returns all facts, optionally filtered by category
	ListFacts(ctx context.Context, category string, limit int) ([]Fact, error)

	// Clear removes all facts
	Clear(ctx context.Context) error

	// Size returns total number of facts
	Size() int
}

// RecallOptions configures memory recall behavior
type RecallOptions struct {
	// MaxMessages limits total messages returned
	MaxMessages int

	// WorkingSize how many recent messages from working memory
	WorkingSize int

	// EpisodicTopK how many similar messages from episodic memory
	EpisodicTopK int

	// SemanticTopK how many relevant facts from semantic memory
	SemanticTopK int

	// MinImportance filters messages below this importance score
	MinImportance float64

	// TimeRange filters messages within time range (optional)
	TimeRange *TimeRange

	// IncludeSummaries whether to include summarized messages
	IncludeSummaries bool

	// Deduplicate removes duplicate or highly similar messages
	Deduplicate bool
}

// SearchFilter for episodic memory queries
type SearchFilter struct {
	Query         string
	MinImportance float64
	TimeRange     *TimeRange
	Tags          []string
	Limit         int
}

// TimeRange represents a time interval
type TimeRange struct {
	Start time.Time
	End   time.Time
}

// Fact represents a piece of knowledge in semantic memory
type Fact struct {
	ID         string                 // Unique identifier
	Content    string                 // Fact content
	Category   string                 // Category (e.g., "preference", "knowledge", "skill")
	Source     string                 // Where this fact came from
	Confidence float64                // Confidence score (0-1)
	CreatedAt  time.Time              // When fact was learned
	UpdatedAt  time.Time              // Last update time
	Metadata   map[string]interface{} // Additional metadata
	Embedding  []float64              // Vector embedding for semantic search
}

// MemoryStats provides statistics about memory usage
type MemoryStats struct {
	// Working memory stats
	WorkingSize     int
	WorkingCapacity int

	// Episodic memory stats
	EpisodicSize   int
	EpisodicOldest time.Time
	EpisodicNewest time.Time

	// Semantic memory stats
	SemanticSize       int
	SemanticCategories []string

	// Overall stats
	TotalMessages     int
	TotalFacts        int
	LastCompression   time.Time
	CompressionCount  int
	AverageImportance float64
}

// MemoryConfig configures the memory system
type MemoryConfig struct {
	// Working memory config
	WorkingCapacity   int    // Max messages in working memory (default: 10)
	SummarizationMode string // "none", "simple", "llm" (default: "simple")

	// Episodic memory config
	EpisodicEnabled   bool    // Enable episodic memory (default: true)
	EpisodicThreshold float64 // Min importance to store (default: 0.5)
	EpisodicMaxSize   int     // Max messages to store (0 = unlimited)

	// Semantic memory config
	SemanticEnabled   bool // Enable semantic memory (default: false)
	SemanticAutoLearn bool // Automatically extract facts (default: false)

	// Compression config
	AutoCompress         bool          // Auto-compress when working memory full (default: true)
	CompressionThreshold int           // Trigger compression at this size (default: WorkingCapacity)
	CompressionInterval  time.Duration // Auto-compress interval (0 = disabled)

	// Importance scoring config
	ImportanceScoring bool // Enable importance scoring (default: true)
	ImportanceWeights ImportanceWeights

	// Deduplication config
	DeduplicationEnabled bool    // Remove duplicates (default: true)
	SimilarityThreshold  float64 // Threshold for considering messages similar (default: 0.9)
}

// ImportanceWeights configures how importance is calculated
type ImportanceWeights struct {
	ExplicitRemember float64 // User said "remember this" (default: 1.0)
	PersonalInfo     float64 // Contains personal information (default: 0.8)
	SuccessfulAction float64 // Led to successful outcome (default: 0.7)
	EmotionalContent float64 // High emotional valence (default: 0.6)
	ReferencedCount  float64 // Referenced multiple times (default: 0.5)
	QuestionAnswer   float64 // Q&A pair (default: 0.4)
	Length           float64 // Message length factor (default: 0.3)
}

// DefaultMemoryConfig returns sensible defaults
func DefaultMemoryConfig() MemoryConfig {
	return MemoryConfig{
		WorkingCapacity:   10,
		SummarizationMode: "simple",

		EpisodicEnabled:   true,
		EpisodicThreshold: 0.5,
		EpisodicMaxSize:   10000,

		SemanticEnabled:   false,
		SemanticAutoLearn: false,

		AutoCompress:         true,
		CompressionThreshold: 10,
		CompressionInterval:  0,

		ImportanceScoring: true,
		ImportanceWeights: DefaultImportanceWeights(),

		DeduplicationEnabled: true,
		SimilarityThreshold:  0.9,
	}
}

// DefaultImportanceWeights returns default importance calculation weights
func DefaultImportanceWeights() ImportanceWeights {
	return ImportanceWeights{
		ExplicitRemember: 1.0,
		PersonalInfo:     0.8,
		SuccessfulAction: 0.7,
		EmotionalContent: 0.6,
		ReferencedCount:  0.5,
		QuestionAnswer:   0.4,
		Length:           0.3,
	}
}

// DefaultRecallOptions returns sensible defaults for recall
func DefaultRecallOptions() RecallOptions {
	return RecallOptions{
		MaxMessages:      20,
		WorkingSize:      7,
		EpisodicTopK:     5,
		SemanticTopK:     3,
		MinImportance:    0.0,
		TimeRange:        nil,
		IncludeSummaries: true,
		Deduplicate:      true,
	}
}
