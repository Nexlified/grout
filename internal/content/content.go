package content

import (
	_ "embed"
	"fmt"
	"math/rand/v2"

	"gopkg.in/yaml.v3"
)

//go:embed data/quotes.yaml
var quotesData []byte

//go:embed data/jokes.yaml
var jokesData []byte

// ContentType represents the type of content (quote or joke)
type ContentType string

const (
	ContentTypeQuote ContentType = "quote"
	ContentTypeJoke  ContentType = "joke"
)

// Manager handles loading and providing quotes/jokes
type Manager struct {
	quotes map[string][]string
	jokes  map[string][]string
}

// NewManager creates a new content manager with preloaded quotes and jokes
func NewManager() (*Manager, error) {
	m := &Manager{
		quotes: make(map[string][]string),
		jokes:  make(map[string][]string),
	}

	// Load quotes
	if err := yaml.Unmarshal(quotesData, &m.quotes); err != nil {
		return nil, fmt.Errorf("failed to parse quotes: %w", err)
	}

	// Load jokes
	if err := yaml.Unmarshal(jokesData, &m.jokes); err != nil {
		return nil, fmt.Errorf("failed to parse jokes: %w", err)
	}

	return m, nil
}

// GetRandom returns a random quote or joke, optionally filtered by category
func (m *Manager) GetRandom(contentType ContentType, category string) (string, error) {
	var data map[string][]string
	var typeName string

	switch contentType {
	case ContentTypeQuote:
		data = m.quotes
		typeName = "quote"
	case ContentTypeJoke:
		data = m.jokes
		typeName = "joke"
	default:
		return "", fmt.Errorf("invalid content type: %s", contentType)
	}

	// If category is specified, use only that category
	if category != "" {
		items, exists := data[category]
		if !exists || len(items) == 0 {
			return "", fmt.Errorf("%s category '%s' not found or empty", typeName, category)
		}
		return items[rand.IntN(len(items))], nil
	}

	// No category specified - collect all items from all categories
	var allItems []string
	for _, items := range data {
		allItems = append(allItems, items...)
	}

	if len(allItems) == 0 {
		return "", fmt.Errorf("no %ss available", typeName)
	}

	return allItems[rand.IntN(len(allItems))], nil
}

// GetCategories returns all available categories for a given content type
func (m *Manager) GetCategories(contentType ContentType) []string {
	var data map[string][]string

	switch contentType {
	case ContentTypeQuote:
		data = m.quotes
	case ContentTypeJoke:
		data = m.jokes
	default:
		return nil
	}

	categories := make([]string, 0, len(data))
	for category := range data {
		categories = append(categories, category)
	}

	return categories
}
