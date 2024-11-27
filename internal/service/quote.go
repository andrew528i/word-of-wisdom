package service

import (
	"context"

	"word-of-wisdom/internal/domain"
	"word-of-wisdom/internal/kit"
)

type quoteService struct {
	quoteRepository domain.QuoteRepository
}

// NewQuoteService creates a new instance of quote service
func NewQuoteService(quoteRepository domain.QuoteRepository) domain.QuoteRepository {
	return &quoteService{
		quoteRepository: quoteRepository,
	}
}

// GetRandom returns a random quote from the storage
func (s *quoteService) GetRandom(ctx context.Context) (*domain.Quote, error) {
	kit.Logger.Info("quote service: get random quote")
	return s.quoteRepository.GetRandom(ctx)
}

// Add stores a new quote in the repository
func (s *quoteService) Add(ctx context.Context, quote *domain.Quote) error {
	kit.Logger.Info("quote service: add quote")
	return s.quoteRepository.Add(ctx, quote)
}
