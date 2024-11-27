# Word of Wisdom TCP Server

A DDoS-protected TCP server that serves wisdom quotes using a Proof of Work (PoW) challenge-response protocol. This project demonstrates a secure way to protect TCP services from DDoS attacks while providing meaningful content to authenticated clients.

## 🌟 Features

- **DDoS Protection**: Implements Proof of Work (PoW) challenge-response protocol
- **TCP Server**: Robust TCP server implementation with proper connection handling
- **Quote Service**: Serves wisdom quotes after successful PoW verification
- **Docker Support**: Containerized setup for both server and client
- **Clean Architecture**: Follows clean architecture principles with clear separation of concerns

## 🏗 Architecture

The project follows clean architecture principles with the following main components:

- `cmd/` - Application entry points
- `internal/` - Internal application code
  - `domain/` - Core business logic and entities
  - `service/` - Application services
  - `repository/` - Data access layer
  - `transport/` - Network communication layer
  - `bootstrap/` - Application bootstrapping
  - `errors/` - Custom error definitions
  - `kit/` - Shared utilities

## 🔒 Proof of Work Implementation

The server implements a Hashcash-like Proof of Work algorithm for DDoS protection. This choice was made for several reasons:

1. **CPU-bound**: The algorithm is CPU-intensive, making it effective against distributed attacks
2. **Asymmetric Workload**: Verification is quick for the server, while solving is computationally expensive for clients
3. **Stateless**: No need to store challenge states, reducing server memory requirements
4. **Proven Technology**: Similar to Bitcoin's mining algorithm, well-tested in production environments

## 🚀 Getting Started

### Prerequisites

- Docker
- Make (optional, for convenience)
- Go 1.23.2 (for local development)

### Running with Docker

1. Start the server and client using Docker Compose:
   ```bash
   docker-compose up --build
   ```

### Manual Build and Run

1. Build the server:
   ```bash
   go build -o server ./cmd/server
   ```

2. Run the server:
   ```bash
   ./server
   ```

## 🛠 Development

### Project Setup

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd word-of-wisdom
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Run tests:
   ```bash
   go test ./...
   ```

## 📝 API Protocol

1. **Client connects** to TCP server
2. **Server sends** PoW challenge
3. **Client solves** the challenge
4. **Client sends** solution
5. **Server verifies** solution
6. If valid, **server sends** wisdom quote
7. Connection closes

## 🐳 Docker Support

The project includes:
- `Dockerfile` - For building the server image
- `docker-compose.yml` - For orchestrating the server and client services

## 🧪 Testing

The project includes comprehensive test coverage:
```bash
go test ./... -cover
```
