package errors

import "errors"

// Quote-related errors
var (
	ErrNoQuotes    = errors.New("no quotes available")
	ErrQuoteExists = errors.New("quote already exists")
)
