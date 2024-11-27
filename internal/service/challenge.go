package service

import (
	"context"
	"crypto/sha256"
	"math/big"
	"time"

	"word-of-wisdom/internal/domain"
	"word-of-wisdom/internal/errors"
	"word-of-wisdom/internal/kit"
)

type challengeService struct {
	challengeRepository domain.ChallengeRepository
	secret              []byte
	complexity          *big.Int
	expirationTime      time.Duration
}

// NewChallengeService creates a new instance of challenge service
func NewChallengeService(
	challengeRepository domain.ChallengeRepository,
	secret []byte,
	complexity *big.Int,
	expirationTime time.Duration,
) domain.ChallengeService {
	return &challengeService{
		challengeRepository: challengeRepository,
		secret:              secret,
		complexity:          complexity,
		expirationTime:      expirationTime,
	}
}

// Generate creates a new challenge with random nonce and configured complexity
func (s *challengeService) Generate(ctx context.Context) (*domain.Challenge, error) {
	kit.Logger.Info("challenge service: generating new challenge")

	// Generate random nonce
	nonce, err := domain.GenerateNonce()
	if err != nil {
		return nil, err
	}

	// Create challenge with configured parameters
	challenge := &domain.Challenge{
		Complexity: new(big.Int).Set(s.complexity),
		Nonce:      nonce,
		ExpiresAt:  time.Now().Add(s.expirationTime),
	}

	// Sign the challenge
	challenge.Sign(s.secret)

	// Store the challenge
	if err := s.challengeRepository.CreateChallenge(ctx, challenge); err != nil {
		kit.Logger.Errorw("failed to store challenge",
			"error", err)
		return nil, err
	}

	kit.Logger.Infow("generated new challenge",
		"id", challenge.ID(),
		"complexity", challenge.Complexity,
		"expires_at", challenge.ExpiresAt)
	return challenge, nil
}

// Verify checks if the provided solution is valid for the challenge
func (s *challengeService) Verify(ctx context.Context, challengeID []byte, solution *big.Int) error {
	kit.Logger.Info("challenge service: verifying solution")

	// 1. Get challenge from repository
	challenge, err := s.challengeRepository.GetChallenge(ctx, challengeID)
	if err != nil {
		kit.Logger.Errorw("failed to get challenge",
			"id", challengeID,
			"error", err)
		return err
	}

	// 2. Verify challenge signature
	if !challenge.VerifySignature(s.secret) {
		kit.Logger.Errorw("challenge signature verification failed",
			"id", challengeID)
		return errors.ErrInvalidChallenge
	}

	// 3. Verify solution
	if !challenge.VerifySolution(solution) {
		kit.Logger.Errorw("invalid solution for challenge",
			"id", challengeID)
		return errors.ErrInvalidSolution
	}

	return nil
}

// Solve attempts to find a solution for the given challenge
func (s *challengeService) Solve(ctx context.Context, challenge *domain.Challenge) (*big.Int, error) {
	kit.Logger.Info("challenge service: solving challenge")

	// 1. Check if challenge has expired
	if time.Now().After(challenge.ExpiresAt) {
		kit.Logger.Errorw("challenge has expired",
			"id", challenge.ID())
		return nil, errors.ErrChallengeExpired
	}

	// 2. Verify challenge signature
	if !challenge.VerifySignature(s.secret) {
		kit.Logger.Errorw("invalid challenge signature",
			"id", challenge.ID())
		return nil, errors.ErrInvalidChallenge
	}

	// 3. Try to find a solution
	solution := big.NewInt(0)
	maxIterations := int64(1000000) // Limit iterations to prevent infinite loop

	for i := int64(0); i < maxIterations; i++ {
		select {
		case <-ctx.Done():
			kit.Logger.Warn("context cancelled while solving challenge")
			return nil, ctx.Err()
		default:
			if challenge.VerifySolution(solution) {
				kit.Logger.Infow("found valid solution",
					"id", challenge.ID(),
					"solution", solution.String())
				return solution, nil
			}
			solution.Add(solution, big.NewInt(1))
		}
	}

	kit.Logger.Errorw("failed to find solution within iteration limit",
		"id", challenge.ID(),
		"max_iterations", maxIterations)
	return nil, errors.ErrSolutionNotFound
}

// calculateSolutionHash computes SHA-256 hash of challenge ID concatenated with solution
func calculateSolutionHash(challengeID []byte, solution *big.Int) []byte {
	hash := sha256.New()
	hash.Write(challengeID)
	hash.Write(solution.Bytes())
	return hash.Sum(nil)
}
