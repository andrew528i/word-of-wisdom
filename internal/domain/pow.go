package domain

import (
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
