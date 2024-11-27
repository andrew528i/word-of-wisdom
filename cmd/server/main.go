package main

import (
	"math/big"
	"os"
	"os/signal"
	"syscall"
	"time"

	"word-of-wisdom/internal/bootstrap"
	"word-of-wisdom/internal/kit"
)

func main() {
	// Create and configure application
	app := bootstrap.New()

	// Configure the application
	app.Config.Address = getEnv("SERVER_ADDRESS", ":8080")
	app.Config.Secret = []byte(getEnv("SECRET_KEY", "your-secret-key-for-signing-challenges"))
	app.Config.Complexity = big.NewInt(getEnvInt("POW_COMPLEXITY", 100000))
	app.Config.ExpirationTime = time.Duration(getEnvInt("CHALLENGE_EXPIRATION_SECONDS", 300)) * time.Second

	// Initialize the application
	if err := app.Init(); err != nil {
		kit.Logger.Fatal("failed to initialize application", err)
	}

	// Start the application in a goroutine
	go func() {
		if err := app.Start(); err != nil {
			kit.Logger.Fatal("failed to start application", err)
		}
	}()

	kit.Logger.Info("application started", "address", app.Config.Address)

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	kit.Logger.Info("shutting down application")

	// Stop the application
	if err := app.Stop(); err != nil {
		kit.Logger.Error("error during shutdown", err)
	}

	kit.Logger.Info("application stopped")
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvInt gets an environment variable as int or returns a default value
func getEnvInt(key string, defaultValue int64) int64 {
	if value, exists := os.LookupEnv(key); exists {
		if n, err := big.NewInt(0).SetString(value, 10); err {
			return n.Int64()
		}
	}
	return defaultValue
}
