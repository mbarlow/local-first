package core

import (
	"crypto/rand"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

// DataProcessor handles core business logic
type DataProcessor struct {
	startTime time.Time
}

// NewDataProcessor creates a new data processor instance
func NewDataProcessor() *DataProcessor {
	return &DataProcessor{
		startTime: time.Now(),
	}
}

// ProcessText performs various text processing operations
func (dp *DataProcessor) ProcessText(input string) (map[string]interface{}, error) {
	if input == "" {
		return nil, fmt.Errorf("empty input provided")
	}

	words := strings.Fields(input)
	sentences := strings.Split(input, ".")

	// Clean up sentences (remove empty ones)
	cleanSentences := make([]string, 0)
	for _, sentence := range sentences {
		if trimmed := strings.TrimSpace(sentence); trimmed != "" {
			cleanSentences = append(cleanSentences, trimmed)
		}
	}

	// Calculate readability metrics
	avgWordsPerSentence := float64(len(words)) / float64(len(cleanSentences))
	if len(cleanSentences) == 0 {
		avgWordsPerSentence = float64(len(words))
	}

	// Word frequency analysis
	wordFreq := make(map[string]int)
	for _, word := range words {
		cleaned := strings.ToLower(strings.Trim(word, ".,!?;:\"'"))
		if cleaned != "" {
			wordFreq[cleaned]++
		}
	}

	// Find most common words
	type wordCount struct {
		Word  string
		Count int
	}

	var wordCounts []wordCount
	for word, count := range wordFreq {
		wordCounts = append(wordCounts, wordCount{Word: word, Count: count})
	}

	sort.Slice(wordCounts, func(i, j int) bool {
		return wordCounts[i].Count > wordCounts[j].Count
	})

	// Take top 5 most common words
	topWords := make([]map[string]interface{}, 0)
	limit := 5
	if len(wordCounts) < limit {
		limit = len(wordCounts)
	}

	for i := 0; i < limit; i++ {
		topWords = append(topWords, map[string]interface{}{
			"word":  wordCounts[i].Word,
			"count": wordCounts[i].Count,
		})
	}

	result := map[string]interface{}{
		"originalLength":      len(input),
		"wordCount":           len(words),
		"sentenceCount":       len(cleanSentences),
		"avgWordsPerSentence": math.Round(avgWordsPerSentence*100) / 100,
		"uniqueWords":         len(wordFreq),
		"topWords":            topWords,
		"readabilityScore":    dp.calculateReadabilityScore(len(words), len(cleanSentences), len(wordFreq)),
		"processed":           true,
		"processingTime":      time.Since(dp.startTime).Milliseconds(),
	}

	return result, nil
}

// CalculateStatistics computes basic statistics for a slice of numbers
func (dp *DataProcessor) CalculateStatistics(numbers []float64) map[string]interface{} {
	if len(numbers) == 0 {
		return map[string]interface{}{
			"error": "no numbers provided",
		}
	}

	// Sort for median calculation
	sorted := make([]float64, len(numbers))
	copy(sorted, numbers)
	sort.Float64s(sorted)

	// Basic calculations
	sum := 0.0
	min := numbers[0]
	max := numbers[0]

	for _, num := range numbers {
		sum += num
		if num < min {
			min = num
		}
		if num > max {
			max = num
		}
	}

	mean := sum / float64(len(numbers))

	// Median calculation
	var median float64
	n := len(sorted)
	if n%2 == 0 {
		median = (sorted[n/2-1] + sorted[n/2]) / 2
	} else {
		median = sorted[n/2]
	}

	// Standard deviation
	variance := 0.0
	for _, num := range numbers {
		variance += math.Pow(num-mean, 2)
	}
	variance /= float64(len(numbers))
	stdDev := math.Sqrt(variance)

	// Quartiles
	q1 := dp.percentile(sorted, 0.25)
	q3 := dp.percentile(sorted, 0.75)

	return map[string]interface{}{
		"count":          len(numbers),
		"sum":            math.Round(sum*100) / 100,
		"mean":           math.Round(mean*100) / 100,
		"median":         math.Round(median*100) / 100,
		"min":            min,
		"max":            max,
		"range":          max - min,
		"standardDev":    math.Round(stdDev*100) / 100,
		"variance":       math.Round(variance*100) / 100,
		"q1":             math.Round(q1*100) / 100,
		"q3":             math.Round(q3*100) / 100,
		"iqr":            math.Round((q3-q1)*100) / 100,
		"processingTime": time.Since(dp.startTime).Milliseconds(),
	}
}

// GenerateID creates different types of identifiers
func (dp *DataProcessor) GenerateID(idType string) string {
	switch idType {
	case "uuid":
		return dp.generateUUID()
	case "short":
		return dp.generateShortID(8)
	case "numeric":
		return fmt.Sprintf("%d", time.Now().UnixNano())
	case "timestamp":
		return time.Now().Format("20060102-150405")
	default:
		return dp.generateShortID(12)
	}
}

// Helper methods

func (dp *DataProcessor) calculateReadabilityScore(wordCount, sentenceCount, uniqueWords int) float64 {
	if sentenceCount == 0 {
		return 0.0
	}

	avgWordsPerSentence := float64(wordCount) / float64(sentenceCount)
	lexicalDiversity := float64(uniqueWords) / float64(wordCount)

	// Simple readability formula (higher is more readable)
	score := 100 - (avgWordsPerSentence * 1.5) + (lexicalDiversity * 50)

	// Clamp between 0 and 100
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return math.Round(score*100) / 100
}

func (dp *DataProcessor) percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}

	index := p * float64(len(sorted)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))

	if lower == upper {
		return sorted[lower]
	}

	weight := index - float64(lower)
	return sorted[lower]*(1-weight) + sorted[upper]*weight
}

func (dp *DataProcessor) generateUUID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)

	// Set version (4) and variant bits
	bytes[6] = (bytes[6] & 0x0f) | 0x40
	bytes[8] = (bytes[8] & 0x3f) | 0x80

	return fmt.Sprintf("%x-%x-%x-%x-%x",
		bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:16])
}

func (dp *DataProcessor) generateShortID(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes := make([]byte, length)
	rand.Read(bytes)

	for i, b := range bytes {
		bytes[i] = charset[b%byte(len(charset))]
	}

	return string(bytes)
}
