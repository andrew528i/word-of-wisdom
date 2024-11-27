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

func TestChallenge_VerifySignature(t *testing.T) {
	baseTime := time.Now().Add(time.Hour) // Future time
	secret := []byte("test secret")

	t.Run("should verify valid non-expired challenge", func(t *testing.T) {
		challenge := &Challenge{
			Complexity: big.NewInt(1000),
			Nonce:      []byte("test nonce"),
			ExpiresAt:  baseTime,
		}
		challenge.Sign(secret)
		assert.True(t, challenge.VerifySignature(secret))
	})

	t.Run("should fail with expired challenge", func(t *testing.T) {
		challenge := &Challenge{
			Complexity: big.NewInt(1000),
			Nonce:      []byte("test nonce"),
			ExpiresAt:  time.Now().Add(-time.Hour), // Past time
		}
		challenge.Sign(secret)
		assert.False(t, challenge.VerifySignature(secret))
	})

	t.Run("should fail with nil signature", func(t *testing.T) {
		challenge := &Challenge{
			Complexity: big.NewInt(1000),
			Nonce:      []byte("test nonce"),
			ExpiresAt:  baseTime,
		}
		assert.False(t, challenge.VerifySignature(secret))
	})

	t.Run("should fail with wrong secret", func(t *testing.T) {
		challenge := &Challenge{
			Complexity: big.NewInt(1000),
			Nonce:      []byte("test nonce"),
			ExpiresAt:  baseTime,
		}
		challenge.Sign(secret)
		wrongSecret := []byte("wrong secret")
		assert.False(t, challenge.VerifySignature(wrongSecret))
	})

	t.Run("should fail with modified complexity", func(t *testing.T) {
		challenge := &Challenge{
			Complexity: big.NewInt(1000),
			Nonce:      []byte("test nonce"),
			ExpiresAt:  baseTime,
		}
		challenge.Sign(secret)
		challenge.Complexity = big.NewInt(500) // Modify after signing
		assert.False(t, challenge.VerifySignature(secret))
	})

	t.Run("should fail with modified nonce", func(t *testing.T) {
		challenge := &Challenge{
			Complexity: big.NewInt(1000),
			Nonce:      []byte("test nonce"),
			ExpiresAt:  baseTime,
		}
		challenge.Sign(secret)
		challenge.Nonce = []byte("modified nonce") // Modify after signing
		assert.False(t, challenge.VerifySignature(secret))
	})

	t.Run("should fail with modified expiration", func(t *testing.T) {
		challenge := &Challenge{
			Complexity: big.NewInt(1000),
			Nonce:      []byte("test nonce"),
			ExpiresAt:  baseTime,
		}
		challenge.Sign(secret)
		challenge.ExpiresAt = baseTime.Add(time.Hour) // Modify after signing
		assert.False(t, challenge.VerifySignature(secret))
	})
}

func TestChallenge_VerifySolution(t *testing.T) {
	baseTime := time.Now().Add(time.Hour) // Future time

	t.Run("should verify valid solution with 1 leading zero", func(t *testing.T) {
		challenge := &Challenge{
			Complexity: big.NewInt(1), // Require 1 leading zero byte
			Nonce:      []byte("test nonce"),
			ExpiresAt:  baseTime,
		}

		// Try solutions until we find a valid one
		solution := big.NewInt(0)
		found := false
		for i := 0; i < 100000 && !found; i++ {
			if challenge.VerifySolution(solution) {
				found = true
				break
			}
			solution.Add(solution, big.NewInt(1))
		}
		assert.True(t, found, "Should find valid solution with 1 leading zero")
	})

	t.Run("should fail with expired challenge", func(t *testing.T) {
		challenge := &Challenge{
			Complexity: big.NewInt(1),
			Nonce:      []byte("test nonce"),
			ExpiresAt:  time.Now().Add(-time.Hour), // Past time
		}
		solution := big.NewInt(42)
		assert.False(t, challenge.VerifySolution(solution))
	})

	t.Run("should fail with zero complexity", func(t *testing.T) {
		challenge := &Challenge{
			Complexity: big.NewInt(0),
			Nonce:      []byte("test nonce"),
			ExpiresAt:  baseTime,
		}
		solution := big.NewInt(42)
		assert.False(t, challenge.VerifySolution(solution))
	})

	t.Run("should fail with negative complexity", func(t *testing.T) {
		challenge := &Challenge{
			Complexity: big.NewInt(-1),
			Nonce:      []byte("test nonce"),
			ExpiresAt:  baseTime,
		}
		solution := big.NewInt(42)
		assert.False(t, challenge.VerifySolution(solution))
	})

	t.Run("should be deterministic for same inputs", func(t *testing.T) {
		challenge := &Challenge{
			Complexity: big.NewInt(1),
			Nonce:      []byte("test nonce"),
			ExpiresAt:  baseTime,
		}
		solution := big.NewInt(42)
		result1 := challenge.VerifySolution(solution)
		result2 := challenge.VerifySolution(solution)
		assert.Equal(t, result1, result2)
	})

	t.Run("should verify solution with 2 leading zeros when required", func(t *testing.T) {
		challenge := &Challenge{
			Complexity: big.NewInt(2), // Require 2 leading zero bytes
			Nonce:      []byte("test nonce"),
			ExpiresAt:  baseTime,
		}

		// Try solutions until we find a valid one
		solution := big.NewInt(0)
		found := false
		for i := 0; i < 200000 && !found; i++ { // More iterations since it's harder
			if challenge.VerifySolution(solution) {
				found = true
				break
			}
			solution.Add(solution, big.NewInt(1))
		}
		assert.True(t, found, "Should find valid solution with 2 leading zeros")
	})
}
