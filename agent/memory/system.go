package memory

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Memory implements the MemorySystem interface
// It orchestrates Working, Episodic, and Semantic memory tiers
type Memory struct {
	working  WorkingMemory
	episodic EpisodicMemory
	semantic SemanticMemory
	config   MemoryConfig

	// Stats tracking
	compressionCount int
	lastCompression  time.Time
	totalMessages    int

	mu sync.RWMutex
}

// New creates a new hierarchical memory system with default configuration
func New() *Memory {
	return NewWithConfig(DefaultMemoryConfig())
}

// NewWithConfig creates a new hierarchical memory system with custom configuration
func NewWithConfig(config MemoryConfig) *Memory {
	return &Memory{
		working:          NewWorkingMemory(config.WorkingCapacity),
		episodic:         NewEpisodicMemory(),
		semantic:         NewSemanticMemory(),
		config:           config,
		compressionCount: 0,
		lastCompression:  time.Time{},
		totalMessages:    0,
	}
}

// Deprecated: Use New() or NewWithConfig() instead
// NewSmartMemory creates a new hierarchical memory system (backward compatibility)
func NewSmartMemory(config MemoryConfig) *Memory {
	return NewWithConfig(config)
}

// Add implements MemorySystem.Add
// Adds a message to the memory system, automatically managing tiers
func (m *Memory) Add(ctx context.Context, msg Message) error {
	m.mu.Lock()

	// Check if compression is needed BEFORE adding
	// This prevents working memory from exceeding capacity
	needsCompressionBefore := m.config.AutoCompress && m.working.Size() >= m.config.WorkingCapacity
	if needsCompressionBefore {
		m.mu.Unlock()
		// Compress outside lock to avoid deadlock
		if err := m.Compress(ctx); err != nil {
			// Don't fail, compression is optional
			_ = err
		}
		m.mu.Lock()
	}

	// Always add to working memory first
	if err := m.working.Add(ctx, msg); err != nil {
		m.mu.Unlock()
		return fmt.Errorf("failed to add to working memory: %w", err)
	}

	m.totalMessages++

	// Calculate importance if enabled
	importance := 0.0
	if m.config.ImportanceScoring {
		importance = m.calculateImportance(msg)
		if msg.Metadata == nil {
			msg.Metadata = make(map[string]interface{})
		}
		msg.Metadata["importance"] = importance
	}

	// Store in episodic memory if important enough
	if m.config.EpisodicEnabled && importance >= m.config.EpisodicThreshold {
		if err := m.episodic.Store(ctx, msg, importance); err != nil {
			// Don't fail the whole operation, just log
			// TODO: Add proper logging
			_ = err
		}
	}

	m.mu.Unlock()

	return nil
}

// Recall implements MemorySystem.Recall
// Retrieves relevant messages from all memory tiers
func (m *Memory) Recall(ctx context.Context, query string, opts RecallOptions) ([]Message, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var allMessages []Message

	// 1. Get recent messages from working memory (hot)
	if opts.WorkingSize > 0 {
		recent, err := m.working.Recent(ctx, opts.WorkingSize)
		if err != nil {
			return nil, fmt.Errorf("failed to get working memory: %w", err)
		}
		allMessages = append(allMessages, recent...)
	}

	// 2. Search episodic memory (warm) if enabled
	if m.config.EpisodicEnabled && opts.EpisodicTopK > 0 {
		filter := SearchFilter{
			Query:         query,
			MinImportance: opts.MinImportance,
			TimeRange:     opts.TimeRange,
			Limit:         opts.EpisodicTopK,
		}

		episodic, err := m.episodic.Search(ctx, filter)
		if err != nil {
			// Don't fail, episodic search is optional
			_ = err
		} else {
			allMessages = append(allMessages, episodic...)
		}
	}

	// 3. Query semantic memory for facts (cold) if enabled
	if m.config.SemanticEnabled && opts.SemanticTopK > 0 {
		facts, err := m.semantic.QueryKnowledge(ctx, query, opts.SemanticTopK)
		if err != nil {
			// Don't fail, semantic search is optional
			_ = err
		} else {
			// Convert facts to messages
			for _, fact := range facts {
				msg := Message{
					Role:      "system",
					Content:   fact.Content,
					Timestamp: fact.CreatedAt,
					Metadata: map[string]interface{}{
						"type":       "fact",
						"category":   fact.Category,
						"confidence": fact.Confidence,
					},
				}
				allMessages = append(allMessages, msg)
			}
		}
	}

	// 4. Deduplicate if enabled
	if opts.Deduplicate {
		allMessages = m.deduplicate(allMessages, m.config.SimilarityThreshold)
	}

	// 5. Limit to max messages
	if opts.MaxMessages > 0 && len(allMessages) > opts.MaxMessages {
		allMessages = allMessages[:opts.MaxMessages]
	}

	return allMessages, nil
}

// Compress implements MemorySystem.Compress
// Compresses working memory by summarizing old messages
func (m *Memory) Compress(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.config.SummarizationMode == "none" {
		return nil
	}

	// Trigger compression on working memory
	summary, compressed, err := m.working.Compress(ctx)
	if err != nil {
		return fmt.Errorf("compression failed: %w", err)
	}

	// Store compressed messages in episodic if enabled
	if m.config.EpisodicEnabled && len(compressed) > 0 {
		importances := make([]float64, len(compressed))
		for i, msg := range compressed {
			if imp, ok := msg.Metadata["importance"].(float64); ok {
				importances[i] = imp
			} else {
				importances[i] = 0.5 // Default importance
			}
		}

		if err := m.episodic.StoreBatch(ctx, compressed, importances); err != nil {
			// Don't fail, just log
			_ = err
		}
	}

	// Add summary back to working memory if it'm not empty
	if summary.Content != "" {
		if err := m.working.Add(ctx, summary); err != nil {
			return fmt.Errorf("failed to add summary: %w", err)
		}
	}

	m.compressionCount++
	m.lastCompression = time.Now()

	return nil
}

// Clear implements MemorySystem.Clear
func (m *Memory) Clear(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.working.Clear(ctx); err != nil {
		return fmt.Errorf("failed to clear working memory: %w", err)
	}

	if m.config.EpisodicEnabled {
		if err := m.episodic.Clear(ctx); err != nil {
			return fmt.Errorf("failed to clear episodic memory: %w", err)
		}
	}

	if m.config.SemanticEnabled {
		if err := m.semantic.Clear(ctx); err != nil {
			return fmt.Errorf("failed to clear semantic memory: %w", err)
		}
	}

	m.totalMessages = 0
	m.compressionCount = 0
	m.lastCompression = time.Time{}

	return nil
}

// Stats implements MemorySystem.Stats
func (m *Memory) Stats(ctx context.Context) MemoryStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := MemoryStats{
		WorkingSize:      m.working.Size(),
		WorkingCapacity:  m.working.Capacity(),
		TotalMessages:    m.totalMessages,
		LastCompression:  m.lastCompression,
		CompressionCount: m.compressionCount,
	}

	if m.config.EpisodicEnabled {
		stats.EpisodicSize = m.episodic.Size()

		// Get oldest/newest timestamps from episodic memory
		if episodicImpl, ok := m.episodic.(*EpisodicMemoryImpl); ok {
			stats.EpisodicOldest = episodicImpl.GetOldestTimestamp()
			stats.EpisodicNewest = episodicImpl.GetNewestTimestamp()
			stats.AverageImportance = episodicImpl.GetAverageImportance()
		}
	}

	if m.config.SemanticEnabled {
		stats.SemanticSize = m.semantic.Size()
		stats.TotalFacts = m.semantic.Size()

		// Get categories from semantic memory
		if semanticImpl, ok := m.semantic.(*SemanticMemoryImpl); ok {
			stats.SemanticCategories = semanticImpl.GetCategories()
		}
	}

	return stats
}

// SetConfig implements MemorySystem.SetConfig
func (m *Memory) SetConfig(config MemoryConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.config = config
	return nil
}

// GetConfig implements MemorySystem.GetConfig
func (m *Memory) GetConfig() MemoryConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.config
}

// calculateImportance calculates importance score for a message
func (m *Memory) calculateImportance(msg Message) float64 {
	weights := m.config.ImportanceWeights
	score := 0.0

	// Check for explicit "remember this" patterns
	content := msg.Content
	if containsRememberKeywords(content) {
		score += weights.ExplicitRemember
	}

	// Check for personal information using enhanced detection from utils
	if hasPersonalInfo(content) {
		score += weights.PersonalInfo
	}

	// Check if it's a Q&A pair
	if isQuestionOrAnswer(content) {
		score += weights.QuestionAnswer
	}

	// Length factor (longer messages might be more important)
	if len(content) > 200 {
		score += weights.Length
	}

	// Check metadata for additional signals
	if msg.Metadata != nil {
		if success, ok := msg.Metadata["successful_action"].(bool); ok && success {
			score += weights.SuccessfulAction
		}
		if refCount, ok := msg.Metadata["reference_count"].(int); ok && refCount > 1 {
			score += weights.ReferencedCount
		}
		if emotional, ok := msg.Metadata["emotional"].(bool); ok && emotional {
			score += weights.EmotionalContent
		}
	}

	// Return raw score (not normalized)
	// Threshold should be configured based on weight values (e.g., 0.7 for high importance)
	return score
}

// deduplicate removes duplicate or highly similar messages
func (m *Memory) deduplicate(messages []Message, threshold float64) []Message {
	if len(messages) <= 1 {
		return messages
	}

	seen := make(map[string]bool)
	unique := make([]Message, 0, len(messages))

	for _, msg := range messages {
		// Simple deduplication by exact content match
		// TODO: Implement semantic similarity comparison
		if !seen[msg.Content] {
			seen[msg.Content] = true
			unique = append(unique, msg)
		}
	}

	return unique
}

// Helper functions for importance calculation

func containsRememberKeywords(content string) bool {
	keywords := []string{
		"remember", "don't forget", "important", "keep in mind",
		"note that", "make sure", "always", "never forget",
	}

	contentLower := toLower(content)
	for _, keyword := range keywords {
		if contains(contentLower, keyword) {
			return true
		}
	}
	return false
}

func containsPersonalInfo(content string) bool {
	// Simple heuristic: contains "my", "I", "me" with specific patterns
	personalPatterns := []string{
		"my name", "I am", "I'm", "my birthday", "I live",
		"my preference", "I like", "I love", "I hate", "I prefer",
	}

	contentLower := toLower(content)
	for _, pattern := range personalPatterns {
		if contains(contentLower, pattern) {
			return true
		}
	}
	return false
}

func isQuestionOrAnswer(content string) bool {
	// Check if it'm a question
	if len(content) > 0 && content[len(content)-1] == '?' {
		return true
	}

	// Check if it starts with question words
	questionWords := []string{"what", "where", "when", "why", "who", "how", "which"}
	contentLower := toLower(content)
	for _, word := range questionWords {
		if startsWithWord(contentLower, word) {
			return true
		}
	}

	return false
}

func contains(s, substr string) bool {
	// Case-insensitive contains using strings package
	s = toLower(s)
	substr = toLower(substr)
	return indexOfSubstring(s, substr) >= 0
}

func toLower(s string) string {
	// Simple ASCII lowercase conversion
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c = c + ('a' - 'A')
		}
		result[i] = c
	}
	return string(result)
}

func indexOfSubstring(s, substr string) int {
	// Simple substring search
	if len(substr) == 0 {
		return 0
	}
	if len(substr) > len(s) {
		return -1
	}

	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

func startsWithWord(s, word string) bool {
	// Case-insensitive word boundary check
	s = toLower(s)
	word = toLower(word)

	if len(s) < len(word) {
		return false
	}

	// Check if starts with word
	if indexOfSubstring(s, word) != 0 {
		return false
	}

	// Check word boundary (space or end of string after word)
	if len(s) == len(word) {
		return true
	}

	nextChar := s[len(word)]
	return nextChar == ' ' || nextChar == '\t' || nextChar == '\n'
}
