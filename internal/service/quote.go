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
func NewQuoteService(quoteRepository domain.QuoteRepository) domain.QuoteService {
	return &quoteService{
		quoteRepository: quoteRepository,
	}
}

// GetRandomQuote returns a random quote from the storage
func (s *quoteService) GetRandomQuote(ctx context.Context) (*domain.Quote, error) {
	kit.Logger.Info("quote service: get random quote")
	return s.quoteRepository.GetRandomQuote(ctx)
}

// CreateQuote stores a new quote in the repository
func (s *quoteService) CreateQuote(ctx context.Context, quote *domain.Quote) error {
	kit.Logger.Info("quote service: create quote")
	return s.quoteRepository.CreateQuote(ctx, quote)
}
