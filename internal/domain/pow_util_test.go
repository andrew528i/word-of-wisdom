package domain

import (
	"bytes"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestChallenge_ID(t *testing.T) {
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	baseChallenge := &Challenge{
		Complexity: big.NewInt(1000),
		Nonce:      []byte("test nonce"),
		ExpiresAt:  baseTime,
	}

	t.Run("should generate 32-byte ID", func(t *testing.T) {
		id := baseChallenge.ID()
		assert.NotNil(t, id)
		assert.Len(t, id, 32, "ID should be 32 bytes (SHA-256)")
	})

	t.Run("should generate same ID for same field values", func(t *testing.T) {
		sameFieldsChallenge := &Challenge{
			Complexity: big.NewInt(0).Set(baseChallenge.Complexity),
			Nonce:      append([]byte(nil), baseChallenge.Nonce...),
			ExpiresAt:  baseChallenge.ExpiresAt,
		}
		assert.True(t, bytes.Equal(baseChallenge.ID(), sameFieldsChallenge.ID()))
	})

	t.Run("should generate different ID for different complexity", func(t *testing.T) {
		differentComplexity := &Challenge{
			Complexity: big.NewInt(2000),
			Nonce:      baseChallenge.Nonce,
			ExpiresAt:  baseChallenge.ExpiresAt,
		}
		assert.False(t, bytes.Equal(baseChallenge.ID(), differentComplexity.ID()))
	})

	t.Run("should generate different ID for different nonce", func(t *testing.T) {
		differentNonce := &Challenge{
			Complexity: baseChallenge.Complexity,
			Nonce:      []byte("different nonce"),
			ExpiresAt:  baseChallenge.ExpiresAt,
		}
		assert.False(t, bytes.Equal(baseChallenge.ID(), differentNonce.ID()))
	})

	t.Run("should generate different ID for different expiration", func(t *testing.T) {
		differentExpiration := &Challenge{
			Complexity: baseChallenge.Complexity,
			Nonce:      baseChallenge.Nonce,
			ExpiresAt:  baseChallenge.ExpiresAt.Add(time.Hour),
		}
		assert.False(t, bytes.Equal(baseChallenge.ID(), differentExpiration.ID()))
	})

	t.Run("should ignore signature field", func(t *testing.T) {
		withSignature := &Challenge{
			Complexity: baseChallenge.Complexity,
			Nonce:      baseChallenge.Nonce,
			ExpiresAt:  baseChallenge.ExpiresAt,
			Signature:  []byte("some signature"),
		}
		assert.True(t, bytes.Equal(baseChallenge.ID(), withSignature.ID()))
	})

	t.Run("should ignore solution field", func(t *testing.T) {
		withSolution := &Challenge{
			Complexity: baseChallenge.Complexity,
			Nonce:      baseChallenge.Nonce,
			ExpiresAt:  baseChallenge.ExpiresAt,
			Solution:   big.NewInt(42),
		}
		assert.True(t, bytes.Equal(baseChallenge.ID(), withSolution.ID()))
	})
}

func TestChallenge_Sign(t *testing.T) {
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	challenge := &Challenge{
		Complexity: big.NewInt(1000),
		Nonce:      []byte("test nonce"),
		ExpiresAt:  baseTime,
	}
	secret := []byte("test secret")

	t.Run("should generate 32-byte signature", func(t *testing.T) {
		challenge.Sign(secret)
		assert.NotNil(t, challenge.Signature)
		assert.Len(t, challenge.Signature, 32, "Signature should be 32 bytes (SHA-256)")
	})

	t.Run("should generate same signature for same inputs", func(t *testing.T) {
		challenge1 := &Challenge{
			Complexity: big.NewInt(1000),
			Nonce:      []byte("test nonce"),
			ExpiresAt:  baseTime,
		}
		challenge2 := &Challenge{
			Complexity: big.NewInt(1000),
			Nonce:      []byte("test nonce"),
			ExpiresAt:  baseTime,
		}

		challenge1.Sign(secret)
		challenge2.Sign(secret)
		assert.True(t, bytes.Equal(challenge1.Signature, challenge2.Signature))
	})

	t.Run("should generate different signature for different secret", func(t *testing.T) {
		challenge1 := &Challenge{
			Complexity: big.NewInt(1000),
			Nonce:      []byte("test nonce"),
			ExpiresAt:  baseTime,
		}
		challenge2 := &Challenge{
			Complexity: big.NewInt(1000),
			Nonce:      []byte("test nonce"),
			ExpiresAt:  baseTime,
		}

		challenge1.Sign([]byte("secret1"))
		challenge2.Sign([]byte("secret2"))
		assert.False(t, bytes.Equal(challenge1.Signature, challenge2.Signature))
	})

	t.Run("should generate different signature for different challenge", func(t *testing.T) {
		challenge1 := &Challenge{
			Complexity: big.NewInt(1000),
			Nonce:      []byte("test nonce 1"),
			ExpiresAt:  baseTime,
		}
		challenge2 := &Challenge{
			Complexity: big.NewInt(1000),
			Nonce:      []byte("test nonce 2"),
			ExpiresAt:  baseTime,
		}

		challenge1.Sign(secret)
		challenge2.Sign(secret)
		assert.False(t, bytes.Equal(challenge1.Signature, challenge2.Signature))
	})
}
