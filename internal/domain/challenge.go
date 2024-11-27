package domain

import (
	"context"
	"math/big"
	"time"
)

// Challenge represents a proof of work challenge sent by the server
type Challenge struct {
	// Complexity defines the required difficulty target for the solution hash
	Complexity *big.Int

	// Nonce is the random value that needs to be used in hash calculation
	Nonce []byte

	// ExpiresAt defines when this challenge becomes invalid
	ExpiresAt time.Time

	// Signature is a cryptographic signature of the challenge parameters
	// to prevent tampering (especially complexity modification)
	Signature []byte

	// Solution is the value that satisfies the proof of work requirement
	// It is nil when the challenge is created and set when solved
	Solution *big.Int
}

// ChallengeService defines operations for managing proof-of-work challenges
type ChallengeService interface {
	// Generate creates a new challenge
	// Returns the signed challenge or error if generation fails
	Generate(ctx context.Context) (*Challenge, error)

	// Verify checks if the provided solution matches the challenge
	// Returns nil if solution is valid, otherwise returns error
	Verify(ctx context.Context, challengeID []byte, solution *big.Int) error

	// Solve attempts to find a solution for the given challenge
	// Returns the solution or error if solving fails
	Solve(ctx context.Context, challenge *Challenge) (*big.Int, error)
}

// ChallengeRepository defines storage operations for challenges
type ChallengeRepository interface {
	// CreateChallenge stores a new challenge
	// Returns error if storage fails
	CreateChallenge(ctx context.Context, challenge *Challenge) error

	// GetChallenge retrieves a challenge by its ID
	// Returns ErrNotFound if challenge doesn't exist
	GetChallenge(ctx context.Context, id []byte) (*Challenge, error)

	// DeleteChallenge removes a challenge by its ID
	// Returns error if deletion fails
	DeleteChallenge(ctx context.Context, id []byte) error
}
