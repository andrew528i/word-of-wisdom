package domain

import (
	"context"
)

// Quote represents a single quote entity with its attributes
type Quote struct {
	// Text contains the actual quote content
	Text string

	// Author represents the person who said or wrote the quote
	Author string

	// Source optionally specifies where the quote comes from (book, speech, etc.)
	Source string
}

// QuoteRepository defines the interface for quote storage operations
type QuoteRepository interface {
	// GetRandom returns a random quote from the storage
	// If no quotes are available, returns errors.ErrNoQuotes
	GetRandom(ctx context.Context) (*Quote, error)

	// Add stores a new quote in the repository
	// Returns errors.ErrQuoteExists if the same quote already exists
	Add(ctx context.Context, quote *Quote) error
}
