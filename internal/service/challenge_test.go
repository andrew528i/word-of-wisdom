package service

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"word-of-wisdom/internal/domain"
	"word-of-wisdom/internal/errors"
	"word-of-wisdom/internal/repository"
)

type ChallengeServiceTestSuite struct {
	suite.Suite
	service     domain.ChallengeService
	repository  domain.ChallengeRepository
	ctx         context.Context
	secret      []byte
	complexity  *big.Int
	expiration  time.Duration
}

func (s *ChallengeServiceTestSuite) SetupTest() {
	s.repository = repository.NewChallengeMemoryRepository()
	s.ctx = context.Background()
	s.secret = []byte("test-secret")
	s.complexity = big.NewInt(1) // Use small complexity for testing
	s.expiration = 5 * time.Minute
	s.service = NewChallengeService(s.repository, s.secret, s.complexity, s.expiration)
}

func TestChallengeService(t *testing.T) {
	suite.Run(t, new(ChallengeServiceTestSuite))
}

func (s *ChallengeServiceTestSuite) TestGenerate_Success() {
	challenge, err := s.service.Generate(s.ctx)
	
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), challenge)
	assert.Equal(s.T(), s.complexity.Int64(), challenge.Complexity.Int64())
	assert.NotNil(s.T(), challenge.Nonce)
	assert.NotNil(s.T(), challenge.Signature)
	assert.True(s.T(), challenge.ExpiresAt.After(time.Now()))
	assert.True(s.T(), challenge.ExpiresAt.Before(time.Now().Add(s.expiration).Add(time.Second)))

	// Verify we can retrieve the challenge from repository
	stored, err := s.repository.GetChallenge(s.ctx, challenge.ID())
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), stored)
}

func (s *ChallengeServiceTestSuite) TestVerify_Success() {
	// Generate a challenge first
	challenge, err := s.service.Generate(s.ctx)
	assert.NoError(s.T(), err)

	// Solve the challenge
	solution := big.NewInt(0)
	found := false
	for i := int64(0); i < 1000000; i++ {
		if challenge.VerifySolution(solution) {
			found = true
			break
		}
		solution.Add(solution, big.NewInt(1))
	}
	assert.True(s.T(), found, "Failed to find a valid solution")
	assert.True(s.T(), challenge.VerifySolution(solution), "Solution verification failed")

	// Verify the solution
	err = s.service.Verify(s.ctx, challenge.ID(), solution)
	assert.NoError(s.T(), err)
}

func (s *ChallengeServiceTestSuite) TestVerify_NotFound() {
	challengeID := []byte("non-existent")
	solution := big.NewInt(42)

	err := s.service.Verify(s.ctx, challengeID, solution)
	
	assert.Error(s.T(), err)
	assert.Equal(s.T(), errors.ErrNotFound, err)
}

func (s *ChallengeServiceTestSuite) TestVerify_InvalidSignature() {
	// Generate a challenge first
	challenge, err := s.service.Generate(s.ctx)
	assert.NoError(s.T(), err)

	// Delete the original challenge
	err = s.repository.DeleteChallenge(s.ctx, challenge.ID())
	assert.NoError(s.T(), err)

	// Tamper with the signature and store the tampered version
	challenge.Signature = []byte("invalid-signature")
	err = s.repository.CreateChallenge(s.ctx, challenge)
	assert.NoError(s.T(), err)

	// Try to verify with any solution
	solution := big.NewInt(42)
	err = s.service.Verify(s.ctx, challenge.ID(), solution)
	
	assert.Error(s.T(), err)
	assert.Equal(s.T(), errors.ErrInvalidChallenge, err)
}

func (s *ChallengeServiceTestSuite) TestVerify_InvalidSolution() {
	// Generate a challenge first
	challenge, err := s.service.Generate(s.ctx)
	assert.NoError(s.T(), err)

	// Try to verify with wrong solution
	wrongSolution := big.NewInt(999)
	err = s.service.Verify(s.ctx, challenge.ID(), wrongSolution)
	
	assert.Error(s.T(), err)
	assert.Equal(s.T(), errors.ErrInvalidSolution, err)
}

func (s *ChallengeServiceTestSuite) TestSolve_Success() {
	// Create a challenge with small complexity for quick test
	challenge := &domain.Challenge{
		Complexity: big.NewInt(1),
		Nonce:      []byte("test-nonce"),
		ExpiresAt:  time.Now().Add(time.Hour),
	}
	challenge.Sign(s.secret)

	solution, err := s.service.Solve(s.ctx, challenge)
	
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), solution)
	assert.True(s.T(), challenge.VerifySolution(solution))
}

func (s *ChallengeServiceTestSuite) TestSolve_ExpiredChallenge() {
	// First generate a valid challenge
	challenge, err := s.service.Generate(s.ctx)
	assert.NoError(s.T(), err)

	// Modify its expiration time to make it expired
	challenge.ExpiresAt = time.Now().Add(-time.Hour)
	err = s.repository.CreateChallenge(s.ctx, challenge)
	assert.NoError(s.T(), err)

	// Try to solve the expired challenge
	solution, err := s.service.Solve(s.ctx, challenge)
	
	assert.Error(s.T(), err)
	assert.Equal(s.T(), errors.ErrChallengeExpired, err)
	assert.Nil(s.T(), solution)
}

func (s *ChallengeServiceTestSuite) TestSolve_InvalidSignature() {
	challenge := &domain.Challenge{
		Complexity: s.complexity,
		Nonce:      []byte("test-nonce"),
		ExpiresAt:  time.Now().Add(time.Hour),
		Signature:  []byte("invalid-signature"),
	}

	solution, err := s.service.Solve(s.ctx, challenge)
	
	assert.Error(s.T(), err)
	assert.Equal(s.T(), errors.ErrInvalidChallenge, err)
	assert.Nil(s.T(), solution)
}
