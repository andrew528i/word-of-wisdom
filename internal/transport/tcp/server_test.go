package tcp

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"math/big"
	"net"
	"testing"
	"time"

	"word-of-wisdom/internal/domain"
	"word-of-wisdom/internal/kit"
	"word-of-wisdom/internal/repository"
	"word-of-wisdom/internal/service"
)

func TestServer(t *testing.T) {
	// Create in-memory repositories
	challengeRepo := repository.NewChallengeMemoryRepository()
	quoteRepo := repository.NewQuoteMemoryRepository()

	// Create services
	challengeService := service.NewChallengeService(
		challengeRepo,
		[]byte("test-secret"),
		big.NewInt(1), // easy complexity for tests
		5*time.Minute,
	)
	quoteService := service.NewQuoteService(quoteRepo)

	// Add test quotes
	testQuotes := []*domain.Quote{
		{
			Text:   "The only way to do great work is to love what you do.",
			Author: "Steve Jobs",
		},
		{
			Text:   "Stay hungry, stay foolish.",
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
	}

	for _, quote := range testQuotes {
		if err := quoteRepo.CreateQuote(context.Background(), quote); err != nil {
			t.Fatalf("Failed to create test quote: %v", err)
		}
	}

	// Add a test quote
	testQuote := &domain.Quote{
		Text:   "Test is important",
		Author: "Gopher",
	}
	if err := quoteRepo.CreateQuote(context.Background(), testQuote); err != nil {
		t.Fatalf("Failed to create test quote: %v", err)
	}

	// Create and start server
	server := NewServer(challengeService, quoteService)
	go func() {
		if err := server.Start(":0"); err != nil {
			t.Errorf("Server failed: %v", err)
		}
	}()
	time.Sleep(100 * time.Millisecond) // wait for server to start

	// Get server address
	addr := server.listener.Addr().String()

	t.Run("Get Challenge and Quote Flow", func(t *testing.T) {
		// Connect to server
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		// Send get challenge request
		if _, err := conn.Write([]byte{0x01}); err != nil { // 0x01 = get challenge
			t.Fatalf("Failed to send challenge request: %v", err)
		}

		// Read response length
		lenBuf := make([]byte, 4)
		if _, err := io.ReadFull(conn, lenBuf); err != nil {
			t.Fatalf("Failed to read response length: %v", err)
		}
		responseLen := int(lenBuf[0])<<24 | int(lenBuf[1])<<16 | int(lenBuf[2])<<8 | int(lenBuf[3])

		// Read response
		responseBuf := make([]byte, responseLen)
		if _, err := io.ReadFull(conn, responseBuf); err != nil {
			t.Fatalf("Failed to read response: %v", err)
		}

		// Parse challenge
		var challenge domain.Challenge
		if err := json.Unmarshal(responseBuf, &challenge); err != nil {
			t.Fatalf("Failed to unmarshal challenge: %v", err)
		}

		// Store original challenge ID before solving
		challengeID := challenge.ID()
		kit.Logger.Info("got challenge", "id", challengeID)

		// Solve challenge using service
		solution, err := challengeService.Solve(context.Background(), &challenge)
		if err != nil {
			t.Fatalf("Failed to solve challenge: %v", err)
		}

		// Connect to the server again
		conn, err = net.Dial("tcp", addr)
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}

		// Send get quote request
		if _, err := conn.Write([]byte{0x02}); err != nil { // 0x02 = get quote
			t.Fatalf("Failed to send quote request: %v", err)
		}

		// Send the original challenge ID (32 bytes)
		if _, err := conn.Write(challengeID); err != nil {
			t.Fatalf("Failed to send challenge ID: %v", err)
		}

		// Send solution (padded to 32 bytes)
		solBytes := make([]byte, 32)
		sBytes := solution.Bytes()
		copy(solBytes[32-len(sBytes):], sBytes) // pad with leading zeros
		if _, err := conn.Write(solBytes); err != nil {
			t.Fatalf("Failed to send solution: %v", err)
		}

		time.Sleep(100 * time.Millisecond)

		// Read response length
		if _, err := io.ReadFull(conn, lenBuf); err != nil {
			t.Fatalf("Failed to read quote response length: %v", err)
		}
		responseLen = int(lenBuf[0])<<24 | int(lenBuf[1])<<16 | int(lenBuf[2])<<8 | int(lenBuf[3])

		// Read response
		responseBuf = make([]byte, responseLen)
		if _, err := io.ReadFull(conn, responseBuf); err != nil {
			t.Fatalf("Failed to read quote response: %v", err)
		}

		// Parse quote
		var quote domain.Quote
		if err := json.Unmarshal(responseBuf, &quote); err != nil {
			t.Fatalf("Failed to unmarshal quote: %v", err)
		}

		kit.Logger.Info("got quote", "quote", quote.Text, "author", quote.Author)

		// Verify quote
		if quote.Text == "" || quote.Author == "" {
			t.Errorf("Got empty quote")
		}
	})

	t.Run("Invalid Challenge ID", func(t *testing.T) {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		// Send get quote request
		if _, err := conn.Write([]byte{0x02}); err != nil {
			t.Fatalf("Failed to send quote request: %v", err)
		}

		// Send invalid challenge ID
		invalidID := bytes.Repeat([]byte{0x00}, 32)
		if _, err := conn.Write(invalidID); err != nil {
			t.Fatalf("Failed to send challenge ID: %v", err)
		}

		// Send dummy solution
		dummySolution := bytes.Repeat([]byte{0x00}, 32)
		if _, err := conn.Write(dummySolution); err != nil {
			t.Fatalf("Failed to send solution: %v", err)
		}

		// Read response length
		lenBuf := make([]byte, 4)
		if _, err := io.ReadFull(conn, lenBuf); err != nil {
			t.Fatalf("Failed to read error response length: %v", err)
		}

		// Expect error response
		responseLen := int(lenBuf[0])<<24 | int(lenBuf[1])<<16 | int(lenBuf[2])<<8 | int(lenBuf[3])
		responseBuf := make([]byte, responseLen)
		if _, err := io.ReadFull(conn, responseBuf); err != nil {
			t.Fatalf("Failed to read error response: %v", err)
		}

		var errResp struct {
			Error string `json:"error"`
		}
		if err := json.Unmarshal(responseBuf, &errResp); err != nil {
			t.Fatalf("Failed to unmarshal error response: %v", err)
		}

		if errResp.Error == "" {
			t.Error("Expected error response, got success")
		}
	})

	t.Run("Invalid Solution", func(t *testing.T) {
		// First get a valid challenge
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		// Send get challenge request
		if _, err := conn.Write([]byte{0x01}); err != nil {
			t.Fatalf("Failed to send challenge request: %v", err)
		}

		// Read response length
		lenBuf := make([]byte, 4)
		if _, err := io.ReadFull(conn, lenBuf); err != nil {
			t.Fatalf("Failed to read response length: %v", err)
		}
		responseLen := int(lenBuf[0])<<24 | int(lenBuf[1])<<16 | int(lenBuf[2])<<8 | int(lenBuf[3])

		// Read response
		responseBuf := make([]byte, responseLen)
		if _, err := io.ReadFull(conn, responseBuf); err != nil {
			t.Fatalf("Failed to read response: %v", err)
		}

		// Parse challenge
		var challenge domain.Challenge
		if err := json.Unmarshal(responseBuf, &challenge); err != nil {
			t.Fatalf("Failed to unmarshal challenge: %v", err)
		}

		challengeID := challenge.ID()
		kit.Logger.Info("got challenge for invalid solution test", "id", challengeID)

		// Connect again for quote request
		conn, err = net.Dial("tcp", addr)
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}

		// Send get quote request
		if _, err := conn.Write([]byte{0x02}); err != nil {
			t.Fatalf("Failed to send quote request: %v", err)
		}

		// Send the challenge ID
		if _, err := conn.Write(challengeID); err != nil {
			t.Fatalf("Failed to send challenge ID: %v", err)
		}

		// Send invalid solution (all zeros)
		invalidSolution := make([]byte, 32)
		if _, err := conn.Write(invalidSolution); err != nil {
			t.Fatalf("Failed to send invalid solution: %v", err)
		}

		time.Sleep(100 * time.Millisecond)

		// Read response length
		if _, err := io.ReadFull(conn, lenBuf); err != nil {
			t.Fatalf("Failed to read error response length: %v", err)
		}
		responseLen = int(lenBuf[0])<<24 | int(lenBuf[1])<<16 | int(lenBuf[2])<<8 | int(lenBuf[3])

		// Read error response
		responseBuf = make([]byte, responseLen)
		if _, err := io.ReadFull(conn, responseBuf); err != nil {
			t.Fatalf("Failed to read error response: %v", err)
		}

		// Parse error response
		var errResponse struct {
			Error string `json:"error"`
		}
		if err := json.Unmarshal(responseBuf, &errResponse); err != nil {
			t.Fatalf("Failed to unmarshal error response: %v", err)
		}

		kit.Logger.Info("got error response", "error", errResponse.Error)

		// Verify we got an invalid solution error
		if errResponse.Error == "" {
			t.Error("Expected error response for invalid solution, got empty error")
		}
	})

	// Cleanup
	if err := server.Stop(); err != nil {
		t.Fatalf("Failed to stop server: %v", err)
	}
}
