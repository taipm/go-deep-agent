# TODO: Fix Integration Test Failures

**Created**: November 10, 2025  
**Priority**: HIGH (Week 1 Day 2 - Consolidation Phase)  
**Status**: Ready to start  
**Goal**: Make all 6 integration tests pass

---

## 🎯 OVERVIEW

Current Status: **4/6 tests failing** (expected - need calibration)

Passing:
- ✅ TestMemoryIntegration_ConcurrentAccess
- ✅ TestMemoryIntegration_FullCycle

Failing:
- ❌ TestMemoryIntegration_WorkingToEpisodic
- ❌ TestMemoryIntegration_ImportanceScoring
- ❌ TestMemoryIntegration_Compression
- ❌ TestMemoryIntegration_VectorRetrieval

---

## 📋 TASK 1: Fix Working Memory Capacity Management

### Issue
```
TestMemoryIntegration_WorkingToEpisodic:
  Expected working size 3, got 4
  Working memory exceeded capacity: 6 > 5
```

### Root Cause Analysis

- [ ] Investigate why working memory exceeds capacity
  ```bash
  # Check working memory implementation
  grep -n "type WorkingMemory" agent/memory/working.go
  grep -n "Add.*ctx.*Message" agent/memory/working.go
  ```

- [ ] Check compression trigger logic
  ```bash
  # Find where compression should happen
  grep -n "AutoCompress" agent/memory/system.go
  grep -n "CompressionThreshold" agent/memory/system.go
  ```

- [ ] Review FIFO eviction implementation
  ```bash
  # Verify FIFO buffer implementation
  grep -n "Compress\|evict\|truncate" agent/memory/working.go
  ```

### Expected Behavior

Working memory should:
1. Never exceed `WorkingCapacity` setting
2. Auto-evict oldest messages when full
3. Trigger compression before adding new message if at capacity

### Fix Strategy

**Option A: Strict Capacity Enforcement** (Recommended)
```go
// In working.go Add() method
func (w *WorkingMemory) Add(ctx context.Context, msg Message) error {
    w.mu.Lock()
    defer w.mu.Unlock()
    
    // Enforce capacity BEFORE adding
    for len(w.messages) >= w.capacity {
        // Remove oldest message
        w.messages = w.messages[1:]
    }
    
    w.messages = append(w.messages, msg)
    return nil
}
```

**Option B: Compression-based Management**
```go
// In system.go Add() method
func (m *Memory) Add(ctx context.Context, msg Message) error {
    // Check capacity BEFORE adding
    if m.working.Size() >= m.config.WorkingCapacity {
        if m.config.AutoCompress {
            if err := m.Compress(ctx); err != nil {
                return err
            }
        }
    }
    
    return m.working.Add(ctx, msg)
}
```

### Implementation Steps

- [ ] **Step 1**: Add capacity check in `working.go::Add()`
  - File: `agent/memory/working.go`
  - Method: `Add(ctx context.Context, msg Message) error`
  - Change: Add FIFO eviction before append

- [ ] **Step 2**: Add pre-add compression in `system.go::Add()`
  - File: `agent/memory/system.go`
  - Method: `Add(ctx context.Context, msg Message) error`
  - Change: Trigger compression BEFORE adding if at capacity

- [ ] **Step 3**: Add unit tests
  ```go
  func TestWorkingMemory_CapacityEnforcement(t *testing.T) {
      wm := NewWorkingMemory(3) // Capacity = 3
      
      // Add 5 messages
      for i := 0; i < 5; i++ {
          wm.Add(ctx, Message{Content: fmt.Sprintf("msg%d", i)})
      }
      
      // Should have exactly 3 messages (last 3)
      if wm.Size() != 3 {
          t.Errorf("Expected size 3, got %d", wm.Size())
      }
      
      // Should have messages 2, 3, 4 (oldest evicted)
      all, _ := wm.All(ctx)
      if all[0].Content != "msg2" {
          t.Error("FIFO eviction failed")
      }
  }
  ```

- [ ] **Step 4**: Run tests to verify
  ```bash
  go test -v -run TestMemoryIntegration_WorkingToEpisodic ./agent/memory/
  go test -v -run TestMemoryIntegration_Compression ./agent/memory/
  ```

### Success Criteria

- ✅ Working memory never exceeds capacity
- ✅ TestMemoryIntegration_WorkingToEpisodic passes
- ✅ TestMemoryIntegration_Compression passes
- ✅ New unit test passes

---

## 📋 TASK 2: Calibrate Importance Scoring

### Issue
```
TestMemoryIntegration_ImportanceScoring:
  Case 1 (Personal information): expected high=true, got importance=0.00
  Message: "My email is john@example.com and my phone is 555-1234"
```

### Root Cause Analysis

- [ ] Find importance calculation logic
  ```bash
  grep -n "calculateImportance" agent/memory/system.go
  ```

- [ ] Check importance weights
  ```bash
  grep -n "ImportanceWeights" agent/memory/interfaces.go
  grep -n "PersonalInfo" agent/memory/
  ```

- [ ] Review keyword detection
  ```bash
  # Find how personal info is detected
  grep -rn "email\|phone\|personal" agent/memory/
  ```

### Expected Behavior

Messages should get high importance (>= 0.6) if they contain:
- ✅ Explicit "remember" keyword → 1.0
- ⚠️ Personal info (email, phone, name) → 0.8 (NOT WORKING)
- ✅ Questions → 0.4
- ✅ Casual chat → 0.3

### Fix Strategy

Need to implement **content-based importance detection**:

```go
// In system.go
func (m *Memory) calculateImportance(msg Message) float64 {
    content := strings.ToLower(msg.Content)
    importance := 0.0
    
    // 1. Explicit remember (highest priority)
    if strings.Contains(content, "remember") {
        importance += m.config.ImportanceWeights.ExplicitRemember // 1.0
    }
    
    // 2. Personal information detection
    if hasPersonalInfo(content) {
        importance += m.config.ImportanceWeights.PersonalInfo // 0.8
    }
    
    // 3. Question/Answer
    if strings.Contains(content, "?") {
        importance += m.config.ImportanceWeights.QuestionAnswer // 0.4
    }
    
    // 4. Length factor (longer = more important)
    if len(msg.Content) > 100 {
        importance += m.config.ImportanceWeights.Length // 0.3
    }
    
    // Normalize to 0-1 range
    if importance > 1.0 {
        importance = 1.0
    }
    
    return importance
}

func hasPersonalInfo(content string) bool {
    // Email pattern
    if matched, _ := regexp.MatchString(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`, content); matched {
        return true
    }
    
    // Phone pattern
    if matched, _ := regexp.MatchString(`\b\d{3}[-.]?\d{3}[-.]?\d{4}\b`, content); matched {
        return true
    }
    
    // Name indicators
    nameIndicators := []string{"my name is", "i'm", "i am", "call me"}
    contentLower := strings.ToLower(content)
    for _, indicator := range nameIndicators {
        if strings.Contains(contentLower, indicator) {
            return true
        }
    }
    
    // Personal keywords
    personalKeywords := []string{"birthday", "allergic", "prefer", "favorite", "address"}
    for _, keyword := range personalKeywords {
        if strings.Contains(contentLower, keyword) {
            return true
        }
    }
    
    return false
}
```

### Implementation Steps

- [ ] **Step 1**: Implement `hasPersonalInfo()` helper
  - File: `agent/memory/system.go`
  - Add regex patterns for email, phone
  - Add keyword detection for names, preferences

- [ ] **Step 2**: Update `calculateImportance()` method
  - File: `agent/memory/system.go`
  - Call `hasPersonalInfo()` to add personal info weight
  - Add other heuristics (question marks, length)

- [ ] **Step 3**: Add unit tests for importance scoring
  ```go
  func TestImportanceScoring_PersonalInfo(t *testing.T) {
      testCases := []struct {
          message  string
          expected float64
      }{
          {"My email is john@example.com", 0.8},
          {"Call me at 555-1234", 0.8},
          {"My name is John", 0.8},
          {"I'm allergic to peanuts", 0.8},
          {"Remember this important fact", 1.0},
          {"What time is it?", 0.4},
          {"Hello", 0.0},
      }
      
      // Test each case
  }
  ```

- [ ] **Step 4**: Run integration test
  ```bash
  go test -v -run TestMemoryIntegration_ImportanceScoring ./agent/memory/
  ```

### Success Criteria

- ✅ Personal info detection works (email, phone, name)
- ✅ All test cases in TestMemoryIntegration_ImportanceScoring pass
- ✅ Importance scores match expected ranges

---

## 📋 TASK 3: Improve Vector Similarity Search

### Issue
```
TestMemoryIntegration_VectorRetrieval:
  Query 'programming languages': expected topic 'programming' in top 3, but not found
  Query 'climate preferences': expected topic 'weather' in top 3, but not found
```

### Root Cause Analysis

- [ ] Check vector embedding implementation
  ```bash
  grep -n "Retrieve\|embedding" agent/memory/episodic.go
  ```

- [ ] Verify similarity calculation
  ```bash
  # Find how similarity is calculated
  grep -n "similarity\|distance\|cosine" agent/memory/
  ```

- [ ] Review retrieval algorithm
  ```bash
  grep -n "func.*Retrieve" agent/memory/episodic.go
  ```

### Expected Behavior

Vector retrieval should:
1. Use semantic similarity (not just keyword matching)
2. Return most relevant messages for query
3. Handle synonyms and related concepts

### Current Implementation Issue

The current `episodic.go` likely uses **simple keyword matching** instead of **semantic embeddings**.

### Fix Strategy

**Option A: Simple TF-IDF Similarity** (Quick fix)
```go
func (e *EpisodicMemory) Retrieve(ctx context.Context, query string, topK int) ([]Message, error) {
    // Calculate TF-IDF similarity between query and each message
    scores := make(map[int]float64)
    
    queryWords := tokenize(query)
    
    for i, msg := range e.messages {
        msgWords := tokenize(msg.Content)
        similarity := calculateTFIDF(queryWords, msgWords)
        scores[i] = similarity
    }
    
    // Sort by score and return top K
    // ...
}
```

**Option B: Use External Embedding Service** (Better, but requires API)
```go
// Add embedding field to Message
type Message struct {
    // ... existing fields
    Embedding []float64 // Vector representation
}

// Generate embeddings when storing
func (e *EpisodicMemory) Store(ctx context.Context, msg Message, importance float64) error {
    // Generate embedding for message content
    embedding, err := e.embedder.Embed(msg.Content)
    if err != nil {
        // Fallback to no embedding
        embedding = nil
    }
    msg.Embedding = embedding
    
    // Store message with embedding
    e.messages = append(e.messages, msg)
    return nil
}

// Retrieve using cosine similarity
func (e *EpisodicMemory) Retrieve(ctx context.Context, query string, topK int) ([]Message, error) {
    queryEmbedding, err := e.embedder.Embed(query)
    if err != nil {
        return nil, err
    }
    
    // Calculate cosine similarity with each message
    scores := make([]struct{
        idx   int
        score float64
    }, len(e.messages))
    
    for i, msg := range e.messages {
        if len(msg.Embedding) > 0 {
            scores[i] = struct{idx int; score float64}{
                idx:   i,
                score: cosineSimilarity(queryEmbedding, msg.Embedding),
            }
        }
    }
    
    // Sort by score descending
    sort.Slice(scores, func(i, j int) bool {
        return scores[i].score > scores[j].score
    })
    
    // Return top K
    results := make([]Message, min(topK, len(scores)))
    for i := 0; i < len(results); i++ {
        results[i] = e.messages[scores[i].idx]
    }
    
    return results, nil
}
```

### Implementation Steps (Option A - Quick Fix)

- [ ] **Step 1**: Implement simple similarity calculation
  - File: `agent/memory/episodic.go`
  - Add `tokenize()` helper
  - Add `calculateSimilarity()` using word overlap

- [ ] **Step 2**: Update `Retrieve()` method
  - Calculate similarity scores for all messages
  - Sort by score (highest first)
  - Return top K

- [ ] **Step 3**: Improve with stemming/lemmatization
  - Handle word variations ("programming" vs "program")
  - Handle synonyms ("weather" vs "climate")

- [ ] **Step 4**: Test with integration test
  ```bash
  go test -v -run TestMemoryIntegration_VectorRetrieval ./agent/memory/
  ```

### Implementation Steps (Option B - Full Solution)

- [ ] **Step 1**: Add embedding interface
  ```go
  type Embedder interface {
      Embed(text string) ([]float64, error)
  }
  ```

- [ ] **Step 2**: Create simple embedder (for tests)
  ```go
  type SimpleEmbedder struct{}
  
  func (s *SimpleEmbedder) Embed(text string) ([]float64, error) {
      // Simple hash-based embedding for testing
      // Or use OpenAI embeddings API
  }
  ```

- [ ] **Step 3**: Update EpisodicMemory to use embeddings

- [ ] **Step 4**: Add embedding generation on Store()

- [ ] **Step 5**: Update Retrieve() to use cosine similarity

### Success Criteria

- ✅ "programming languages" query returns programming-related messages
- ✅ "climate preferences" query returns weather-related messages
- ✅ TestMemoryIntegration_VectorRetrieval passes
- ✅ Semantic similarity works (not just keywords)

---

## 📋 TASK 4: Add Helper Functions

### Shared Utilities Needed

Create `agent/memory/utils.go`:

```go
package memory

import (
    "math"
    "regexp"
    "strings"
)

// tokenize splits text into words
func tokenize(text string) []string {
    text = strings.ToLower(text)
    // Remove punctuation
    reg := regexp.MustCompile(`[^\w\s]`)
    text = reg.ReplaceAllString(text, "")
    // Split on whitespace
    return strings.Fields(text)
}

// cosineSimilarity calculates cosine similarity between two vectors
func cosineSimilarity(a, b []float64) float64 {
    if len(a) != len(b) || len(a) == 0 {
        return 0.0
    }
    
    var dotProduct, normA, normB float64
    for i := range a {
        dotProduct += a[i] * b[i]
        normA += a[i] * a[i]
        normB += b[i] * b[i]
    }
    
    normA = math.Sqrt(normA)
    normB = math.Sqrt(normB)
    
    if normA == 0 || normB == 0 {
        return 0.0
    }
    
    return dotProduct / (normA * normB)
}

// jaccardSimilarity calculates Jaccard similarity (word overlap)
func jaccardSimilarity(words1, words2 []string) float64 {
    set1 := make(map[string]bool)
    set2 := make(map[string]bool)
    
    for _, w := range words1 {
        set1[w] = true
    }
    for _, w := range words2 {
        set2[w] = true
    }
    
    // Intersection
    intersection := 0
    for w := range set1 {
        if set2[w] {
            intersection++
        }
    }
    
    // Union
    union := len(set1) + len(set2) - intersection
    
    if union == 0 {
        return 0.0
    }
    
    return float64(intersection) / float64(union)
}
```

### Implementation Steps

- [ ] **Step 1**: Create `agent/memory/utils.go`
- [ ] **Step 2**: Implement helper functions
- [ ] **Step 3**: Add unit tests for utils
- [ ] **Step 4**: Use in importance scoring and retrieval

---

## 📋 TASK 5: Update Tests to Match Implementation

### Adjust Test Expectations

Some test failures might be due to **unrealistic expectations**. Review each test:

- [ ] **TestMemoryIntegration_WorkingToEpisodic**
  - Review expected working size vs capacity
  - Adjust if test expectations are wrong

- [ ] **TestMemoryIntegration_ImportanceScoring**
  - Verify importance thresholds are reasonable
  - Adjust expected scores if needed

- [ ] **TestMemoryIntegration_Compression**
  - Check if compression trigger logic is correct
  - Update test expectations if implementation is correct

- [ ] **TestMemoryIntegration_VectorRetrieval**
  - Consider if simple keyword matching is acceptable
  - Update test if full semantic search is future work

### Strategy

For each failing test:
1. **Understand** what the test expects
2. **Verify** if expectation is realistic
3. **Fix implementation** OR **adjust test**
4. **Document** design decision

---

## 📋 EXECUTION PLAN

### Day 1 (Today - Nov 10)
- [x] Create this TODO list
- [ ] Analyze all 4 failing tests
- [ ] Prioritize fixes

### Day 2 (Nov 11)
- [ ] Task 1: Fix Working Memory Capacity (2-3 hours)
- [ ] Task 4: Add Helper Functions (1 hour)
- [ ] Run tests, verify 2 tests fixed

### Day 3 (Nov 12)
- [ ] Task 2: Calibrate Importance Scoring (2-3 hours)
- [ ] Run tests, verify 1 more test fixed

### Day 4 (Nov 13)
- [ ] Task 3: Improve Vector Retrieval (3-4 hours)
  - Option A (quick) or Option B (full)
- [ ] Run all tests, verify all pass

### Day 5 (Nov 14)
- [ ] Task 5: Review and polish
- [ ] Add documentation
- [ ] Final testing
- [ ] Commit and push

---

## ✅ DEFINITION OF DONE

- [ ] All 6 integration tests pass
- [ ] Test coverage increased (>75%)
- [ ] Code is documented
- [ ] No linting warnings
- [ ] Commit message explains changes
- [ ] Update V0.6.0_CONSOLIDATION_PLAN.md with progress

---

## 📝 NOTES

### Design Decisions to Document

1. **Working Memory Capacity**:
   - Decision: Strict enforcement vs. soft limit?
   - Chosen: [TBD]
   - Rationale: [TBD]

2. **Importance Scoring**:
   - Decision: Simple heuristics vs. ML-based?
   - Chosen: Heuristics (for now)
   - Rationale: Simple, fast, no external dependencies

3. **Vector Retrieval**:
   - Decision: Keyword matching vs. embeddings?
   - Chosen: [TBD - Option A or B]
   - Rationale: [TBD]

### Future Improvements (Post v0.6.0)

- [ ] Add OpenAI embeddings integration (optional)
- [ ] Implement more sophisticated importance scoring
- [ ] Add user feedback for importance calibration
- [ ] Support custom importance functions

---

**Last Updated**: November 10, 2025  
**Status**: Ready to execute  
**Estimated Time**: 3-4 days
