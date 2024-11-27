package domain

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
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
