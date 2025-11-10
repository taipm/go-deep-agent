package memory

import (
	"math"
	"testing"
)

// TestTokenize tests the tokenize function
func TestTokenize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Simple sentence",
			input:    "Hello World",
			expected: []string{"hello", "world"},
		},
		{
			name:     "With punctuation",
			input:    "Hello, World! How are you?",
			expected: []string{"hello", "world", "how", "are", "you"},
		},
		{
			name:     "Mixed case",
			input:    "Programming in Go",
			expected: []string{"programming", "in", "go"},
		},
		{
			name:     "Empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "Only punctuation",
			input:    "!@#$%^&*()",
			expected: []string{},
		},
		{
			name:     "Multiple spaces",
			input:    "hello   world",
			expected: []string{"hello", "world"},
		},
		{
			name:     "Numbers",
			input:    "code 123 test",
			expected: []string{"code", "123", "test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tokenize(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("tokenize(%q) length = %d, want %d", tt.input, len(result), len(tt.expected))
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("tokenize(%q)[%d] = %q, want %q", tt.input, i, result[i], tt.expected[i])
				}
			}
		})
	}
}

// TestCosineSimilarity tests the cosineSimilarity function
func TestCosineSimilarity(t *testing.T) {
	tests := []struct {
		name     string
		a        []float64
		b        []float64
		expected float64
		epsilon  float64
	}{
		{
			name:     "Identical vectors",
			a:        []float64{1.0, 2.0, 3.0},
			b:        []float64{1.0, 2.0, 3.0},
			expected: 1.0,
			epsilon:  0.0001,
		},
		{
			name:     "Orthogonal vectors",
			a:        []float64{1.0, 0.0},
			b:        []float64{0.0, 1.0},
			expected: 0.0,
			epsilon:  0.0001,
		},
		{
			name:     "Opposite vectors",
			a:        []float64{1.0, 2.0, 3.0},
			b:        []float64{-1.0, -2.0, -3.0},
			expected: -1.0,
			epsilon:  0.0001,
		},
		{
			name:     "Different lengths",
			a:        []float64{1.0, 2.0},
			b:        []float64{1.0, 2.0, 3.0},
			expected: 0.0,
			epsilon:  0.0001,
		},
		{
			name:     "Empty vectors",
			a:        []float64{},
			b:        []float64{},
			expected: 0.0,
			epsilon:  0.0001,
		},
		{
			name:     "Zero vector",
			a:        []float64{0.0, 0.0, 0.0},
			b:        []float64{1.0, 2.0, 3.0},
			expected: 0.0,
			epsilon:  0.0001,
		},
		{
			name:     "Similar vectors",
			a:        []float64{1.0, 2.0, 3.0},
			b:        []float64{2.0, 4.0, 6.0},
			expected: 1.0,
			epsilon:  0.0001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cosineSimilarity(tt.a, tt.b)
			if math.Abs(result-tt.expected) > tt.epsilon {
				t.Errorf("cosineSimilarity(%v, %v) = %f, want %f", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// TestJaccardSimilarity tests the jaccardSimilarity function
func TestJaccardSimilarity(t *testing.T) {
	tests := []struct {
		name     string
		words1   []string
		words2   []string
		expected float64
		epsilon  float64
	}{
		{
			name:     "Identical sets",
			words1:   []string{"hello", "world"},
			words2:   []string{"hello", "world"},
			expected: 1.0,
			epsilon:  0.0001,
		},
		{
			name:     "No overlap",
			words1:   []string{"hello", "world"},
			words2:   []string{"foo", "bar"},
			expected: 0.0,
			epsilon:  0.0001,
		},
		{
			name:     "Partial overlap",
			words1:   []string{"hello", "world", "test"},
			words2:   []string{"hello", "foo", "bar"},
			expected: 0.2, // 1 intersection, 5 union
			epsilon:  0.0001,
		},
		{
			name:     "Empty sets",
			words1:   []string{},
			words2:   []string{},
			expected: 0.0,
			epsilon:  0.0001,
		},
		{
			name:     "One empty set",
			words1:   []string{"hello"},
			words2:   []string{},
			expected: 0.0,
			epsilon:  0.0001,
		},
		{
			name:     "Subset",
			words1:   []string{"hello", "world"},
			words2:   []string{"hello", "world", "test"},
			expected: 2.0 / 3.0, // 2 intersection, 3 union
			epsilon:  0.0001,
		},
		{
			name:     "Duplicates in input",
			words1:   []string{"hello", "hello", "world"},
			words2:   []string{"hello", "world"},
			expected: 1.0, // Sets ignore duplicates
			epsilon:  0.0001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := jaccardSimilarity(tt.words1, tt.words2)
			if math.Abs(result-tt.expected) > tt.epsilon {
				t.Errorf("jaccardSimilarity(%v, %v) = %f, want %f", tt.words1, tt.words2, result, tt.expected)
			}
		})
	}
}

// TestHasPersonalInfo tests the hasPersonalInfo function
func TestHasPersonalInfo(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected bool
	}{
		// Email tests
		{
			name:     "Valid email",
			content:  "Contact me at john@example.com",
			expected: true,
		},
		{
			name:     "Email with numbers",
			content:  "My email is user123@test.org",
			expected: true,
		},
		{
			name:     "No personal info",
			content:  "The sky looks beautiful",
			expected: false,
		},

		// Phone tests
		{
			name:     "Phone with dashes",
			content:  "Call me at 555-123-4567",
			expected: true,
		},
		{
			name:     "Phone with parentheses",
			content:  "My number is (555) 123-4567",
			expected: true,
		},
		{
			name:     "Phone with dots",
			content:  "Contact: 555.123.4567",
			expected: true,
		},
		{
			name:     "Short phone",
			content:  "Call 555-1234",
			expected: true,
		},
		{
			name:     "International phone",
			content:  "Phone: +1-555-123-4567",
			expected: true,
		},

		// Name indicator tests
		{
			name:     "My name is",
			content:  "Hi, my name is John",
			expected: true,
		},
		{
			name:     "I'm",
			content:  "I'm Alice",
			expected: true,
		},
		{
			name:     "I am",
			content:  "I am Bob Smith",
			expected: true,
		},
		{
			name:     "Call me",
			content:  "Call me Mike",
			expected: true,
		},

		// Personal keyword tests
		{
			name:     "Birthday",
			content:  "My birthday is tomorrow",
			expected: true,
		},
		{
			name:     "Allergic",
			content:  "I'm allergic to peanuts",
			expected: true,
		},
		{
			name:     "Allergy",
			content:  "I have an allergy to cats",
			expected: true,
		},
		{
			name:     "Prefer",
			content:  "I prefer vegetarian food",
			expected: true,
		},
		{
			name:     "Favorite",
			content:  "My favorite color is blue",
			expected: true,
		},
		{
			name:     "Live in",
			content:  "I live in New York",
			expected: true,
		},
		{
			name:     "Born in",
			content:  "I was born in 1990",
			expected: true,
		},
		{
			name:     "Years old",
			content:  "I am 25 years old",
			expected: true,
		},
		{
			name:     "Work at",
			content:  "I work at Google",
			expected: true,
		},

		// Edge cases
		{
			name:     "Empty string",
			content:  "",
			expected: false,
		},
		{
			name:     "Generic text",
			content:  "The weather is nice today",
			expected: false,
		},
		{
			name:     "Case insensitive",
			content:  "MY NAME IS JOHN",
			expected: true,
		},
		{
			name:     "Multiple indicators",
			content:  "My name is John and my birthday is tomorrow",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasPersonalInfo(tt.content)
			if result != tt.expected {
				t.Errorf("hasPersonalInfo(%q) = %v, want %v", tt.content, result, tt.expected)
			}
		})
	}
}

// TestExpandKeywords tests the expandKeywords function
func TestExpandKeywords(t *testing.T) {
	tests := []struct {
		name          string
		words         []string
		shouldContain []string
	}{
		{
			name:          "Programming expansion",
			words:         []string{"programming"},
			shouldContain: []string{"programming", "coding", "development", "software"},
		},
		{
			name:          "Food expansion",
			words:         []string{"food"},
			shouldContain: []string{"food", "cuisine", "dish", "meal"},
		},
		{
			name:          "Weather expansion",
			words:         []string{"weather"},
			shouldContain: []string{"weather", "climate", "season"},
		},
		{
			name:          "No expansion",
			words:         []string{"unknown", "word"},
			shouldContain: []string{"unknown", "word"},
		},
		{
			name:          "Multiple words with expansion",
			words:         []string{"programming", "weather"},
			shouldContain: []string{"programming", "coding", "weather", "climate"},
		},
		{
			name:          "Empty input",
			words:         []string{},
			shouldContain: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandKeywords(tt.words)

			// Convert result to map for easy lookup
			resultMap := make(map[string]bool)
			for _, word := range result {
				resultMap[word] = true
			}

			// Check all expected words are present
			for _, expected := range tt.shouldContain {
				if !resultMap[expected] {
					t.Errorf("expandKeywords(%v) missing expected word %q", tt.words, expected)
				}
			}
		})
	}
}

// TestTextSimilarity tests the textSimilarity function
func TestTextSimilarity(t *testing.T) {
	tests := []struct {
		name     string
		text1    string
		text2    string
		minScore float64
		maxScore float64
	}{
		{
			name:     "Identical texts",
			text1:    "Hello world",
			text2:    "Hello world",
			minScore: 0.9,
			maxScore: 1.0,
		},
		{
			name:     "Similar with expansion",
			text1:    "I love programming",
			text2:    "I enjoy coding",
			minScore: 0.1, // Should have some similarity due to keyword expansion
			maxScore: 1.0,
		},
		{
			name:     "Weather and climate",
			text1:    "The weather is nice",
			text2:    "The climate is good",
			minScore: 0.1, // Should have some similarity due to expansion
			maxScore: 1.0,
		},
		{
			name:     "Completely different",
			text1:    "I love pizza",
			text2:    "The car is fast",
			minScore: 0.0,
			maxScore: 0.3,
		},
		{
			name:     "Empty strings",
			text1:    "",
			text2:    "",
			minScore: 0.0,
			maxScore: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := textSimilarity(tt.text1, tt.text2)
			if result < tt.minScore || result > tt.maxScore {
				t.Errorf("textSimilarity(%q, %q) = %f, want between %f and %f",
					tt.text1, tt.text2, result, tt.minScore, tt.maxScore)
			}
		})
	}
}
