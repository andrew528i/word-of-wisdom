package tcp

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net"
	"time"

	"word-of-wisdom/internal/domain"
	"word-of-wisdom/internal/kit"
)

const (
	cmdGetChallenge byte = 0x01
	cmdGetQuote     byte = 0x02

	readTimeout  = 10 * time.Second
	writeTimeout = 10 * time.Second
)

// Server handles TCP connections and processes client requests
type Server struct {
	listener   net.Listener
	challenges domain.ChallengeService
	quotes     domain.QuoteService
	shutdownCh chan struct{}
}

// NewServer creates a new TCP server instance
func NewServer(challenges domain.ChallengeService, quotes domain.QuoteService) *Server {
	return &Server{
		challenges: challenges,
		quotes:     quotes,
		shutdownCh: make(chan struct{}),
	}
}

// Start begins listening for connections on the specified address
func (s *Server) Start(address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	s.listener = listener
	kit.Logger.Info("server started", "address", address)

	for {
		select {
		case <-s.shutdownCh:
			return nil
		default:
			conn, err := listener.Accept()
			if err != nil {
				kit.Logger.Error("failed to accept connection", err)
				continue
			}
			go s.handleConnection(conn)
		}
	}
}

// Stop gracefully shuts down the server
func (s *Server) Stop() error {
	close(s.shutdownCh)
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	kit.Logger.Info("new connection", "remote_addr", conn.RemoteAddr())

	// Read command
	if err := conn.SetReadDeadline(time.Now().Add(readTimeout)); err != nil {
		kit.Logger.Error("failed to set read deadline", err)
		return
	}

	cmd := make([]byte, 1)
	if _, err := io.ReadFull(conn, cmd); err != nil {
		kit.Logger.Error("failed to read command", err)
		return
	}

	switch cmd[0] {
	case cmdGetChallenge:
		s.handleGetChallenge(conn)
	case cmdGetQuote:
		s.handleGetQuote(conn)
	default:
		kit.Logger.Error("unknown command", fmt.Errorf("command: %d", cmd[0]))
		s.writeError(conn, fmt.Errorf("unknown command: %d", cmd[0]))
	}
}

func (s *Server) handleGetChallenge(conn net.Conn) {
	challenge, err := s.challenges.Generate(context.Background())
	if err != nil {
		s.writeError(conn, err)
		return
	}

	kit.Logger.Info("generated challenge", "id", challenge.ID())
	s.writeResponse(conn, challenge)
}

func (s *Server) handleGetQuote(conn net.Conn) {
	// Set read deadline for the entire operation
	if err := conn.SetReadDeadline(time.Now().Add(readTimeout)); err != nil {
		kit.Logger.Error("failed to set read deadline", err)
		return
	}

	// Challenge ID is SHA-256 hash (32 bytes)
	id := make([]byte, 32)
	if _, err := io.ReadFull(conn, id); err != nil {
		kit.Logger.Error("failed to read challenge ID", err)
		s.writeError(conn, fmt.Errorf("failed to read challenge ID: %w", err))
		return
	}

	// Solution is 256-bit number (32 bytes)
	solBytes := make([]byte, 32)
	if _, err := io.ReadFull(conn, solBytes); err != nil {
		kit.Logger.Error("failed to read solution", err)
		s.writeError(conn, fmt.Errorf("failed to read solution: %w", err))
		return
	}

	// Verify solution
	if err := s.challenges.Verify(context.Background(), id, new(big.Int).SetBytes(solBytes)); err != nil {
		kit.Logger.Error("invalid solution", err, "challenge_id", fmt.Sprintf("%x", id))
		s.writeError(conn, err)
		return
	}

	// Get random quote
	quote, err := s.quotes.GetRandomQuote(context.Background())
	if err != nil {
		kit.Logger.Error("failed to get random quote", err)
		s.writeError(conn, err)
		return
	}

	kit.Logger.Info("sending quote", "challenge_id", fmt.Sprintf("%x", id))
	s.writeResponse(conn, quote)
}

func (s *Server) writeResponse(conn net.Conn, data interface{}) {
	// Set write deadline for the entire operation
	if err := conn.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
		kit.Logger.Error("failed to set write deadline", err)
		return
	}

	response, err := json.Marshal(data)
	if err != nil {
		kit.Logger.Error("failed to marshal response", err)
		s.writeError(conn, fmt.Errorf("failed to marshal response: %w", err))
		return
	}

	// Write response length
	lenBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBytes, uint32(len(response)))
	n, err := conn.Write(lenBytes)
	if err != nil || n != 4 {
		kit.Logger.Error("failed to write response length", "error", err, "bytes_written", n)
		return
	}

	// Write response data
	n, err = conn.Write(response)
	if err != nil || n != len(response) {
		kit.Logger.Error("failed to write response", "error", err, "bytes_written", n, "expected_bytes", len(response))
		return
	}

	// Reset write deadline after writing
	if err := conn.SetWriteDeadline(time.Time{}); err != nil {
		kit.Logger.Error("failed to reset write deadline", err)
		return
	}
}

func (s *Server) writeError(conn net.Conn, err error) {
	response := struct {
		Error string `json:"error"`
	}{
		Error: err.Error(),
	}
	s.writeResponse(conn, response)
}
