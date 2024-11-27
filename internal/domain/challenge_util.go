package domain

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"math/big"
	"time"
)

// ID generates a unique identifier for the challenge by hashing its fields
func (s *Challenge) ID() []byte {
	// Create a buffer to hold all fields that contribute to the ID
	buf := new(bytes.Buffer)

	// Write complexity as bytes
	complexityBytes := s.Complexity.Bytes()
	_ = binary.Write(buf, binary.BigEndian, int64(len(complexityBytes)))
	buf.Write(complexityBytes)

	// Write nonce
	_ = binary.Write(buf, binary.BigEndian, int64(len(s.Nonce)))
	buf.Write(s.Nonce)

	// Write expiration time as Unix timestamp
	_ = binary.Write(buf, binary.BigEndian, s.ExpiresAt.Unix())

	// Calculate SHA-256 hash of all fields
	hash := sha256.Sum256(buf.Bytes())
	return hash[:]
}

// Sign calculates and sets the signature for the challenge using the provided secret
// The signature is a SHA-256 hash of the challenge ID concatenated with the secret
func (s *Challenge) Sign(secret []byte) {
	// Create a buffer to hold ID and secret
	buf := new(bytes.Buffer)

	// Write challenge ID
	id := s.ID()
	buf.Write(id)

	// Write secret
	buf.Write(secret)

	// Calculate signature
	signature := sha256.Sum256(buf.Bytes())
	s.Signature = signature[:]
}

// VerifySignature checks if the challenge signature is valid using the provided secret
// Returns true if signature is valid and challenge has not expired, false otherwise
func (s *Challenge) VerifySignature(secret []byte) bool {
	// Check if signature exists
	if s.Signature == nil {
		return false
	}

	// Check if challenge has expired
	if time.Now().After(s.ExpiresAt) {
		return false
	}

	// Create a buffer to hold ID and secret
	buf := new(bytes.Buffer)

	// Write challenge ID
	id := s.ID()
	buf.Write(id)

	// Write secret
	buf.Write(secret)

	// Calculate expected signature
	expectedSignature := sha256.Sum256(buf.Bytes())

	// Compare with stored signature
	return bytes.Equal(expectedSignature[:], s.Signature)
}

// VerifySolution checks if the provided solution satisfies the proof-of-work requirement
// The hash of (challenge_id + solution) must have enough leading zeros based on complexity
func (s *Challenge) VerifySolution(solution *big.Int) bool {
	// Check if challenge has expired
	if time.Now().After(s.ExpiresAt) {
		return false
	}

	// Calculate hash of challenge ID and solution
	hash := sha256.New()
	hash.Write(s.ID())
	hash.Write(solution.Bytes())
	hashBytes := hash.Sum(nil)

	// Convert complexity to number of leading zero bytes required
	requiredZeros := s.Complexity.Int64()
	if requiredZeros <= 0 {
		return false
	}

	// Check leading zeros in the hash
	for i := int64(0); i < requiredZeros; i++ {
		if hashBytes[i] != 0 {
			return false
		}
	}
	return true
}

// GenerateNonce creates a cryptographically secure random nonce
func GenerateNonce() ([]byte, error) {
	nonce := make([]byte, 32)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	return nonce, nil
}
