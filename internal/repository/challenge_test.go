package repository

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"word-of-wisdom/internal/domain"
	"word-of-wisdom/internal/errors"
)

type ChallengeRepositoryTestSuite struct {
	suite.Suite
	repo domain.ChallengeRepository
	ctx  context.Context
}

func (s *ChallengeRepositoryTestSuite) SetupTest() {
	s.repo = NewChallengeMemoryRepository()
	s.ctx = context.Background()
}

func TestChallengeRepository(t *testing.T) {
	suite.Run(t, new(ChallengeRepositoryTestSuite))
}

func (s *ChallengeRepositoryTestSuite) TestGetNonExistentChallenge() {
	challenge, err := s.repo.GetChallenge(s.ctx, []byte("non-existent"))
	assert.ErrorIs(s.T(), err, errors.ErrNotFound)
	assert.Nil(s.T(), challenge)
}

func (s *ChallengeRepositoryTestSuite) TestCreateAndGetChallenge() {
	// Create test challenge
	challenge := &domain.Challenge{
		Complexity: big.NewInt(100),
		Nonce:     []byte("test-nonce"),
		ExpiresAt: time.Now().Add(time.Hour),
		Signature: []byte("test-signature"),
		Solution:  big.NewInt(42),
	}

	// Store the challenge
	err := s.repo.CreateChallenge(s.ctx, challenge)
	assert.NoError(s.T(), err)

	// Retrieve the challenge
	stored, err := s.repo.GetChallenge(s.ctx, challenge.ID())
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), stored)

	// Verify all fields match
	assert.Equal(s.T(), challenge.Complexity.Int64(), stored.Complexity.Int64())
	assert.Equal(s.T(), challenge.Nonce, stored.Nonce)
	assert.Equal(s.T(), challenge.ExpiresAt.Unix(), stored.ExpiresAt.Unix())
	assert.Equal(s.T(), challenge.Signature, stored.Signature)
	assert.Equal(s.T(), challenge.Solution.Int64(), stored.Solution.Int64())
}

func (s *ChallengeRepositoryTestSuite) TestCreateDuplicate() {
	challenge := &domain.Challenge{
		Complexity: big.NewInt(100),
		Nonce:     []byte("test-nonce"),
		ExpiresAt: time.Now().Add(time.Hour),
		Signature: []byte("test-signature"),
	}

	// First creation should succeed
	err := s.repo.CreateChallenge(s.ctx, challenge)
	assert.NoError(s.T(), err)

	// Second creation should fail
	err = s.repo.CreateChallenge(s.ctx, challenge)
	assert.ErrorIs(s.T(), err, errors.ErrChallengeExists)
}

func (s *ChallengeRepositoryTestSuite) TestDeleteChallenge() {
	challenge := &domain.Challenge{
		Complexity: big.NewInt(100),
		Nonce:     []byte("test-nonce"),
		ExpiresAt: time.Now().Add(time.Hour),
		Signature: []byte("test-signature"),
	}

	// Create challenge
	err := s.repo.CreateChallenge(s.ctx, challenge)
	assert.NoError(s.T(), err)

	// Delete challenge
	err = s.repo.DeleteChallenge(s.ctx, challenge.ID())
	assert.NoError(s.T(), err)

	// Verify challenge is gone
	stored, err := s.repo.GetChallenge(s.ctx, challenge.ID())
	assert.ErrorIs(s.T(), err, errors.ErrNotFound)
	assert.Nil(s.T(), stored)

	// Delete non-existent challenge should fail
	err = s.repo.DeleteChallenge(s.ctx, []byte("non-existent"))
	assert.ErrorIs(s.T(), err, errors.ErrNotFound)
}

func (s *ChallengeRepositoryTestSuite) TestChallengeImmutability() {
	originalChallenge := &domain.Challenge{
		Complexity: big.NewInt(100),
		Nonce:     []byte("test-nonce"),
		ExpiresAt: time.Now().Add(time.Hour),
		Signature: []byte("test-signature"),
		Solution:  big.NewInt(42),
	}

	// Get the ID before any modifications
	id := originalChallenge.ID()

	// Store the challenge
	err := s.repo.CreateChallenge(s.ctx, originalChallenge)
	assert.NoError(s.T(), err)

	// Modify the original challenge
	originalChallenge.Complexity.SetInt64(200)
	originalChallenge.Nonce = []byte("modified-nonce")
	originalChallenge.Signature = []byte("modified-signature")
	originalChallenge.Solution.SetInt64(84)

	// Get the stored challenge using the original ID
	stored, err := s.repo.GetChallenge(s.ctx, id)
	assert.NoError(s.T(), err)

	// Verify the stored challenge was not modified
	assert.Equal(s.T(), int64(100), stored.Complexity.Int64())
	assert.Equal(s.T(), []byte("test-nonce"), stored.Nonce)
	assert.Equal(s.T(), []byte("test-signature"), stored.Signature)
	assert.Equal(s.T(), int64(42), stored.Solution.Int64())
}
