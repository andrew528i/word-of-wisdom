package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"word-of-wisdom/internal/domain"
	"word-of-wisdom/internal/errors"
)

type QuoteRepositoryTestSuite struct {
	suite.Suite
	repo domain.QuoteRepository
	ctx  context.Context
}

func (s *QuoteRepositoryTestSuite) SetupTest() {
	s.repo = NewQuoteMemoryRepository()
	s.ctx = context.Background()
}

func TestQuoteRepository(t *testing.T) {
	suite.Run(t, new(QuoteRepositoryTestSuite))
}

func (s *QuoteRepositoryTestSuite) TestGetRandomEmpty() {
	// When repository is empty, should return ErrNoQuotes
	quote, err := s.repo.GetRandomQuote(s.ctx)
	assert.ErrorIs(s.T(), err, errors.ErrNoQuotes)
	assert.Nil(s.T(), quote)
}

func (s *QuoteRepositoryTestSuite) TestCreateAndGetRandom() {
	testCases := []struct {
		name    string
		quotes  []*domain.Quote
		wantErr error
	}{
		{
			name: "single quote",
			quotes: []*domain.Quote{
				{
					Text:   "Test quote 1",
					Author: "Author 1",
				},
			},
			wantErr: nil,
		},
		{
			name: "multiple quotes",
			quotes: []*domain.Quote{
				{
					Text:   "Test quote 2",
					Author: "Author 2",
				},
				{
					Text:   "Test quote 3",
					Author: "Author 3",
				},
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Add quotes
			for _, q := range tc.quotes {
				err := s.repo.CreateQuote(s.ctx, q)
				assert.NoError(s.T(), err)
			}

			// Get random quote multiple times to ensure randomness
			quotesMap := make(map[string]bool)
			for i := 0; i < 10; i++ {
				quote, err := s.repo.GetRandomQuote(s.ctx)
				assert.NoError(s.T(), err)
				assert.NotNil(s.T(), quote)
				quotesMap[quote.Text] = true
			}

			// Verify that we got at least one quote
			assert.Greater(s.T(), len(quotesMap), 0)
		})
	}
}

func (s *QuoteRepositoryTestSuite) TestCreateDuplicate() {
	testQuote := &domain.Quote{
		Text:   "Test quote",
		Author: "Author",
	}

	// First add should succeed
	err := s.repo.CreateQuote(s.ctx, testQuote)
	assert.NoError(s.T(), err)

	// Second add with same content should fail
	err = s.repo.CreateQuote(s.ctx, testQuote)
	assert.ErrorIs(s.T(), err, errors.ErrQuoteExists)

	// Add with different content should succeed
	newQuote := &domain.Quote{
		Text:   "Different quote",
		Author: "Author",
	}
	err = s.repo.CreateQuote(s.ctx, newQuote)
	assert.NoError(s.T(), err)
}

func (s *QuoteRepositoryTestSuite) TestQuoteImmutability() {
	originalQuote := &domain.Quote{
		Text:   "Original text",
		Author: "Original author",
	}

	// Add the quote
	err := s.repo.CreateQuote(s.ctx, originalQuote)
	assert.NoError(s.T(), err)

	// Modify the original quote
	originalQuote.Text = "Modified text"
	originalQuote.Author = "Modified author"

	// Get the quote from repository
	storedQuote, err := s.repo.GetRandomQuote(s.ctx)
	assert.NoError(s.T(), err)

	// Verify the stored quote wasn't modified
	assert.Equal(s.T(), "Original text", storedQuote.Text)
	assert.Equal(s.T(), "Original author", storedQuote.Author)
}
