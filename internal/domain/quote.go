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

// QuoteService defines operations for managing quotes
type QuoteService interface {
	// GetRandomQuote returns a random quote from the storage
	// If no quotes are available, returns errors.ErrNoQuotes
	GetRandomQuote(ctx context.Context) (*Quote, error)

	// CreateQuote stores a new quote in the repository
	// Returns errors.ErrQuoteExists if the same quote already exists
	CreateQuote(ctx context.Context, quote *Quote) error
}

// QuoteRepository defines the interface for quote storage operations
type QuoteRepository interface {
	// GetRandomQuote returns a random quote from the storage
	// If no quotes are available, returns errors.ErrNoQuotes
	GetRandomQuote(ctx context.Context) (*Quote, error)

	// CreateQuote stores a new quote in the repository
	// Returns errors.ErrQuoteExists if the same quote already exists
	CreateQuote(ctx context.Context, quote *Quote) error
}
