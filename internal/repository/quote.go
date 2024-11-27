package repository

import (
	"context"
	"math/rand"
	"sync"

	"word-of-wisdom/internal/domain"
	"word-of-wisdom/internal/errors"
	"word-of-wisdom/internal/kit"
)

// quoteMemoryRepository implements domain.QuoteRepository interface with in-memory storage
type quoteMemoryRepository struct {
	sync.RWMutex
	quotes []*domain.Quote
}

// NewQuoteMemoryRepository creates a new in-memory quote repository
func NewQuoteMemoryRepository() domain.QuoteRepository {
	kit.Logger.Info("initializing in-memory quote repository")
	return &quoteMemoryRepository{
		quotes: make([]*domain.Quote, 0),
	}
}

// GetRandom returns a random quote from the storage
func (r *quoteMemoryRepository) GetRandom(ctx context.Context) (*domain.Quote, error) {
	r.RLock()
	defer r.RUnlock()

	if len(r.quotes) == 0 {
		kit.Logger.Error("failed to get random quote: no quotes available")
		return nil, errors.ErrNoQuotes
	}

	randomIndex := rand.Intn(len(r.quotes))
	quote := r.quotes[randomIndex]
	kit.Logger.Infow("retrieved random quote",
		"author", quote.Author)
	return quote, nil
}

// Add stores a new quote in the repository
func (r *quoteMemoryRepository) Add(ctx context.Context, quote *domain.Quote) error {
	r.Lock()
	defer r.Unlock()

	// Check for duplicates by comparing text and author
	for _, existing := range r.quotes {
		if existing.Text == quote.Text && existing.Author == quote.Author {
			kit.Logger.Errorw("failed to add quote: already exists",
				"author", quote.Author)
			return errors.ErrQuoteExists
		}
	}

	// Make a copy of the quote to prevent external modifications
	quoteCopy := &domain.Quote{
		Text:   quote.Text,
		Author: quote.Author,
		Source: quote.Source,
	}

	r.quotes = append(r.quotes, quoteCopy)
	kit.Logger.Infow("added new quote",
		"author", quote.Author)
	return nil
}
