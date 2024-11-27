package repository

import (
	"context"
	"encoding/hex"
	"math/big"
	"sync"

	"word-of-wisdom/internal/domain"
	"word-of-wisdom/internal/errors"
	"word-of-wisdom/internal/kit"
)

// challengeMemoryRepository implements domain.ChallengeRepository interface with in-memory storage
type challengeMemoryRepository struct {
	sync.RWMutex
	challenges map[string]*domain.Challenge // key is hex-encoded challenge ID
}

// NewChallengeMemoryRepository creates a new in-memory challenge repository
func NewChallengeMemoryRepository() domain.ChallengeRepository {
	kit.Logger.Info("initializing in-memory challenge repository")
	return &challengeMemoryRepository{
		challenges: make(map[string]*domain.Challenge),
	}
}

// getHexID returns hex encoded ID for the challenge
func getHexID(challenge *domain.Challenge) string {
	return hex.EncodeToString(challenge.ID())
}

// CreateChallenge stores a new challenge
func (r *challengeMemoryRepository) CreateChallenge(ctx context.Context, challenge *domain.Challenge) error {
	r.Lock()
	defer r.Unlock()

	// Convert ID to hex string for use as map key
	id := getHexID(challenge)

	// Check if challenge with this ID already exists
	if _, exists := r.challenges[id]; exists {
		kit.Logger.Errorw("failed to create challenge: already exists",
			"id", id)
		return errors.ErrChallengeExists
	}

	// Make a deep copy of the challenge to prevent external modifications
	challengeCopy := &domain.Challenge{
		Complexity: new(big.Int).Set(challenge.Complexity),
		Nonce:      make([]byte, len(challenge.Nonce)),
		ExpiresAt:  challenge.ExpiresAt,
		Signature:  make([]byte, len(challenge.Signature)),
	}
	copy(challengeCopy.Nonce, challenge.Nonce)
	copy(challengeCopy.Signature, challenge.Signature)
	if challenge.Solution != nil {
		challengeCopy.Solution = new(big.Int).Set(challenge.Solution)
	}

	r.challenges[id] = challengeCopy
	kit.Logger.Infow("created new challenge",
		"id", id,
		"expires_at", challenge.ExpiresAt)
	return nil
}

// GetChallenge retrieves a challenge by its ID
func (r *challengeMemoryRepository) GetChallenge(ctx context.Context, id []byte) (*domain.Challenge, error) {
	r.RLock()
	defer r.RUnlock()

	hexID := hex.EncodeToString(id)
	challenge, exists := r.challenges[hexID]
	if !exists {
		kit.Logger.Errorw("failed to get challenge: not found",
			"id", hexID)
		return nil, errors.ErrNotFound
	}

	// Return a copy to prevent external modifications
	challengeCopy := &domain.Challenge{
		Complexity: new(big.Int).Set(challenge.Complexity),
		Nonce:      make([]byte, len(challenge.Nonce)),
		ExpiresAt:  challenge.ExpiresAt,
		Signature:  make([]byte, len(challenge.Signature)),
	}
	copy(challengeCopy.Nonce, challenge.Nonce)
	copy(challengeCopy.Signature, challenge.Signature)
	if challenge.Solution != nil {
		challengeCopy.Solution = new(big.Int).Set(challenge.Solution)
	}

	kit.Logger.Infow("retrieved challenge",
		"id", hexID,
		"expires_at", challenge.ExpiresAt)
	return challengeCopy, nil
}

// DeleteChallenge removes a challenge by its ID
func (r *challengeMemoryRepository) DeleteChallenge(ctx context.Context, id []byte) error {
	r.Lock()
	defer r.Unlock()

	hexID := hex.EncodeToString(id)
	if _, exists := r.challenges[hexID]; !exists {
		kit.Logger.Errorw("failed to delete challenge: not found",
			"id", hexID)
		return errors.ErrNotFound
	}

	delete(r.challenges, hexID)
	kit.Logger.Infow("deleted challenge",
		"id", hexID)
	return nil
}
