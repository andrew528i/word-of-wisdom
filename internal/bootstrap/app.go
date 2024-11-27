package bootstrap

import (
	"context"
	"math/big"
	"time"

	"word-of-wisdom/internal/domain"
	"word-of-wisdom/internal/kit"
	"word-of-wisdom/internal/repository"
	"word-of-wisdom/internal/service"
	"word-of-wisdom/internal/transport/tcp"
)

// App represents the main application container
type App struct {
	// Configuration
	Config struct {
		Address          string        // TCP server address
		Secret          []byte        // Secret for challenge signing
		Complexity      *big.Int      // PoW complexity
		ExpirationTime  time.Duration // Challenge expiration time
	}

	// Repositories
	ChallengeRepository domain.ChallengeRepository
	QuoteRepository     domain.QuoteRepository

	// Services
	ChallengeService domain.ChallengeService
	QuoteService     domain.QuoteService

	// Server
	Server *tcp.Server
}

// New creates a new App instance
func New() *App {
	return new(App)
}

// Init initializes the application with the provided configuration
func (a *App) Init() error {
	// Set default configuration if not provided
	if a.Config.Address == "" {
		a.Config.Address = ":8080"
	}
	if a.Config.Secret == nil {
		a.Config.Secret = []byte("default-secret-key")
	}
	if a.Config.Complexity == nil {
		a.Config.Complexity = big.NewInt(100000) // Default complexity
	}
	if a.Config.ExpirationTime == 0 {
		a.Config.ExpirationTime = 5 * time.Minute
	}

	// Initialize repositories
	a.ChallengeRepository = repository.NewChallengeMemoryRepository()
	a.QuoteRepository = repository.NewQuoteMemoryRepository()

	// Initialize services
	a.ChallengeService = service.NewChallengeService(
		a.ChallengeRepository,
		a.Config.Secret,
		a.Config.Complexity,
		a.Config.ExpirationTime,
	)
	a.QuoteService = service.NewQuoteService(a.QuoteRepository)

	// Initialize server
	a.Server = tcp.NewServer(a.ChallengeService, a.QuoteService)

	kit.Logger.Info("application initialized",
		"address", a.Config.Address,
		"complexity", a.Config.Complexity,
		"expiration_time", a.Config.ExpirationTime,
	)

	return nil
}

// Start starts the application
func (a *App) Start() error {
	// Add some default quotes if repository is empty
	if err := a.addDefaultQuotes(); err != nil {
		return err
	}

	kit.Logger.Info("starting server", "address", a.Config.Address)
	return a.Server.Start(a.Config.Address)
}

// Stop gracefully stops the application
func (a *App) Stop() error {
	kit.Logger.Info("stopping server")
	return a.Server.Stop()
}

// addDefaultQuotes adds some default quotes to the repository
func (a *App) addDefaultQuotes() error {
	quotes := []*domain.Quote{
		{
			Text:   "The only way to do great work is to love what you do.",
			Author: "Steve Jobs",
		},
		{
			Text:   "Talk is cheap. Show me the code.",
			Author: "Linus Torvalds",
		},
		{
			Text:   "Programming isn't about what you know; it's about what you can figure out.",
			Author: "Chris Pine",
		},
		{
			Text:   "The best error message is the one that never shows up.",
			Author: "Thomas Fuchs",
		},
		{
			Text:   "First solve the problem. Then write the code.",
			Author: "John Johnson",
		},
	}

	for _, quote := range quotes {
		if err := a.QuoteRepository.CreateQuote(context.Background(), quote); err != nil {
			return err
		}
	}

	return nil
}
