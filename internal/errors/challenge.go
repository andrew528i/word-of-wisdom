package errors

import "errors"

// Challenge-related errors
var (
	ErrNotFound          = errors.New("not found")
	ErrChallengeExists   = errors.New("challenge already exists")
	ErrChallengeExpired  = errors.New("challenge has expired")
	ErrInvalidSolution   = errors.New("invalid solution")
	ErrNoSolutionFound   = errors.New("no solution found")
	ErrInvalidChallenge  = errors.New("invalid challenge")
	ErrSolutionNotFound  = errors.New("solution not found")
)
