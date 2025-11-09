package memory

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// SemanticMemoryImpl implements the SemanticMemory interface
// Stores long-term facts and knowledge
type SemanticMemoryImpl struct {
	facts map[string]Fact // factID -> Fact
	mu    sync.RWMutex
}

// NewSemanticMemory creates a new semantic memory
func NewSemanticMemory() *SemanticMemoryImpl {
	return &SemanticMemoryImpl{
		facts: make(map[string]Fact),
	}
}

// StoreFact implements SemanticMemory.StoreFact
func (s *SemanticMemoryImpl) StoreFact(ctx context.Context, fact Fact) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate ID if not provided
	if fact.ID == "" {
		fact.ID = generateFactID()
	}

	// Set timestamps
	if fact.CreatedAt.IsZero() {
		fact.CreatedAt = time.Now()
	}
	fact.UpdatedAt = time.Now()

	s.facts[fact.ID] = fact
	return nil
}

// QueryKnowledge implements SemanticMemory.QueryKnowledge
// Simple implementation: returns all facts
// TODO: Implement semantic search with vector embeddings
func (s *SemanticMemoryImpl) QueryKnowledge(ctx context.Context, query string, limit int) ([]Fact, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Fact, 0, limit)
	count := 0

	for _, fact := range s.facts {
		result = append(result, fact)
		count++
		if count >= limit {
			break
		}
	}

	return result, nil
}

// UpdateFact implements SemanticMemory.UpdateFact
func (s *SemanticMemoryImpl) UpdateFact(ctx context.Context, factID string, fact Fact) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.facts[factID]; !exists {
		return fmt.Errorf("fact not found: %s", factID)
	}

	fact.ID = factID
	fact.UpdatedAt = time.Now()
	s.facts[factID] = fact

	return nil
}

// DeleteFact implements SemanticMemory.DeleteFact
func (s *SemanticMemoryImpl) DeleteFact(ctx context.Context, factID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.facts[factID]; !exists {
		return fmt.Errorf("fact not found: %s", factID)
	}

	delete(s.facts, factID)
	return nil
}

// ListFacts implements SemanticMemory.ListFacts
func (s *SemanticMemoryImpl) ListFacts(ctx context.Context, category string, limit int) ([]Fact, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Fact, 0)
	count := 0

	for _, fact := range s.facts {
		// Filter by category if specified
		if category != "" && fact.Category != category {
			continue
		}

		result = append(result, fact)
		count++
		if limit > 0 && count >= limit {
			break
		}
	}

	return result, nil
}

// Clear implements SemanticMemory.Clear
func (s *SemanticMemoryImpl) Clear(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.facts = make(map[string]Fact)
	return nil
}

// Size implements SemanticMemory.Size
func (s *SemanticMemoryImpl) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.facts)
}

// generateFactID generates a unique ID for a fact
func generateFactID() string {
	return fmt.Sprintf("fact_%d", time.Now().UnixNano())
}
