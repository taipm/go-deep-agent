package memory

import (
	"math"
	"regexp"
	"strings"
)

// tokenize splits text into lowercase words
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

// jaccardSimilarity calculates Jaccard similarity (word overlap) between two texts
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

// hasPersonalInfo detects if content contains personal information
// including email, phone, name patterns, and personal keywords
func hasPersonalInfo(content string) bool {
	// Email pattern
	emailPattern := regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}\b`)
	if emailPattern.MatchString(content) {
		return true
	}

	// Phone patterns (various formats)
	// 555-1234, (555) 123-4567, 555.123.4567, etc.
	phonePatterns := []string{
		`\b\d{3}[-.\s]?\d{3}[-.\s]?\d{4}\b`,                     // 555-123-4567
		`\(\d{3}\)\s*\d{3}[-.\s]?\d{4}\b`,                       // (555) 123-4567
		`\b\d{3}[-.\s]?\d{4}\b`,                                 // 555-1234
		`\+\d{1,3}[-.\s]?\d{1,4}[-.\s]?\d{1,4}[-.\s]?\d{1,9}\b`, // International
	}
	for _, pattern := range phonePatterns {
		phoneRe := regexp.MustCompile(pattern)
		if phoneRe.MatchString(content) {
			return true
		}
	}

	contentLower := strings.ToLower(content)

	// Name indicators
	nameIndicators := []string{
		"my name is",
		"i'm ",
		"i am ",
		"call me ",
		"this is ",
	}
	for _, indicator := range nameIndicators {
		if strings.Contains(contentLower, indicator) {
			return true
		}
	}

	// Personal information keywords
	personalKeywords := []string{
		"birthday",
		"allergic",
		"allergy",
		"prefer",
		"favorite",
		"favourite",
		"address",
		"live in",
		"live at",
		"born in",
		"born on",
		"age is",
		"years old",
		"work at",
		"work for",
		"employed",
		"graduated",
		"studied",
		"my email",
		"my phone",
		"contact me",
	}

	for _, keyword := range personalKeywords {
		if strings.Contains(contentLower, keyword) {
			return true
		}
	}

	return false
}

// textSimilarity calculates similarity between two texts using Jaccard similarity
// with keyword expansion for better semantic matching
func textSimilarity(text1, text2 string) float64 {
	words1 := tokenize(text1)
	words2 := tokenize(text2)

	// Expand words with related terms for better matching
	expanded1 := expandKeywords(words1)
	expanded2 := expandKeywords(words2)

	return jaccardSimilarity(expanded1, expanded2)
}

// expandKeywords adds related terms to improve semantic matching
func expandKeywords(words []string) []string {
	// Simple keyword expansion dictionary
	expansions := map[string][]string{
		"programming": {"coding", "development", "software", "code", "program"},
		"coding":      {"programming", "development", "software", "code"},
		"language":    {"languages", "lang"},
		"languages":   {"language", "lang"},
		"food":        {"cuisine", "dish", "meal", "eating"},
		"dietary":     {"food", "eating", "allergy", "allergic"},
		"allergic":    {"allergy", "dietary", "food"},
		"weather":     {"climate", "season", "sunny", "rainy", "temperature"},
		"climate":     {"weather", "season", "temperature"},
	}

	expanded := make([]string, 0, len(words)*2)
	added := make(map[string]bool)

	// Add original words
	for _, word := range words {
		if !added[word] {
			expanded = append(expanded, word)
			added[word] = true
		}
	}

	// Add related terms
	for _, word := range words {
		if related, ok := expansions[word]; ok {
			for _, rel := range related {
				if !added[rel] {
					expanded = append(expanded, rel)
					added[rel] = true
				}
			}
		}
	}

	return expanded
}
