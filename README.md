# Ethereum Block Reward API

REST API in Go for querying Ethereum consensus and execution layer data, specifically block rewards and sync committee duties.

## Features

- **Block Reward Endpoint**: Retrieves block rewards and determines if a block was built by MEV builders
- **Sync Duties Endpoint**: Returns validator public keys with sync committee duties for a given slot
- Clean architecture with minimal abstractions

## Architecture

### Framework: Gin
- **Why Gin**: High performance, minimal overhead, excellent middleware support, and widely adopted
- Clean routing syntax and built-in JSON validation

### Configuration Management
- **Dedicated config module**: Centralized configuration with `.env` file support via `godotenv`
- **Environment variable priority**: System env vars override `.env` file values
- **Type-safe configuration**: Validates and converts config values to appropriate types
- **Default values**: Sensible defaults for all configuration options

### Structure
- **Modular design**: Separate packages for handlers and Ethereum client logic
- **Single responsibility**: Each component has a clear, focused purpose
- **Error handling**: Consistent JSON error responses with appropriate HTTP status codes

### Ethereum Integration
- Direct HTTP/JSON-RPC communication without heavy dependencies
- Supports both beacon chain (consensus) and execution layer queries
- Efficient big number handling for Wei to Gwei conversions

## Prerequisites

- Go 1.24 or higher
- Access to an Ethereum RPC endpoint

## Installation

1. Clone the repository:
```bash
git clone git@github.com:alatras/ethereum-validator-api.git
cd ethereum-validator-api
```

2. Copy the environment variables example file:
```bash
cp .env.example .env
```

3. Edit `.env` file with your configuration
```bash
# Edit with your preferred editor
nano .env
```

4. Install dependencies:
```bash
go mod download
```

5. Build the application:
```bash
go build -o ethereum-validator-api
```

## Configuration

Create a `.env` file in the project root (see `.env.example` for reference):

```bash
# Server Configuration
PORT=8080

# Ethereum RPC Configuration
ETH_RPC_URL=https://your-rpc-endpoint.com
```

## Running the API

### Development
```bash
# Development
go run main.go

# Production
./ethereum-validator-api
```

The server will start on port 8080 by default.

## Using the Makefile

You can use the Makefile for common tasks like building, running, and testing the application:

```bash
# Initial setup (creates .env file if not exists and downloads dependencies)
make setup

# Build the application
make build

# Run the application
make run

# Run tests
make test
```

### Development Tools

```bash
# Run with hot-reload (requires air to be installed)
make dev

# Format code
make fmt

# Run linter (requires golangci-lint)
make lint

# Clean build artifacts
make clean
```

### Docker Support

```bash
# Build Docker image
make docker-build

# Run Docker container
make docker-run

# Start Docker container
make up

# Stop Docker container
make down
```

The Docker container exposes port 8080 and uses the ETH_RPC_URL environment variable for configuration.


## API Endpoints

### 1. Get Block Reward

Retrieves the block reward for a given slot and determines if it was built by an MEV builder.

**Endpoint**: `GET /blockreward/{slot}`

**Parameters**:
- `slot` (path parameter): Ethereum slot number

**Response**:
```json
{
  "status": "MEV",  // or "VANILLA"
  "reward": 123456  // reward in GWEI (integer)
}
```

**Error Responses**:
- `400`: Slot is in the future or invalid slot number
- `404`: Slot exists but block is missing
- `500`: Internal server error

**Example**:
```bash
# Get reward for slot 7847950
curl http://localhost:8080/blockreward/7847950

# Response:
{
  "status": "VANILLA",
  "reward": 32518
}
```

### 2. Get Sync Committee Duties

Returns validator public keys that have sync committee duties for a given slot.

**Endpoint**: `GET /syncduties/{slot}`

**Parameters**:
- `slot` (path parameter): Ethereum slot number

**Response**:
```json
[
  "0x93247f2209abcacf57b75a51dafae777f9dd38bc7053d1af526f220a7489a6d3a2753e5f3e8b1cfe39b56f43611df74a",
  "0xa572cbea904d67468808c8eb50a9450c9721db309128012543902d0ac358a62ae28f75bb8f1c7c42c39a8c5529bf0f4e"
]
```

**Error Responses**:
- `400`: Slot is too far in the future (>1 epoch beyond current)
- `404`: Slot exists but no duties found
- `500`: Internal server error

**Example**:
```bash
# Get sync duties for slot 7847950
curl http://localhost:8080/syncduties/7847950

# Response (mock data if real data unavailable):
[
  "0x93247f2209abcacf57b75a51dafae777f9dd38bc7053d1af526f220a7489a6d3a2753e5f3e8b1cfe39b56f43611df74a",
  "0xa572cbea904d67468808c8eb50a9450c9721db309128012543902d0ac358a62ae28f75bb8f1c7c42c39a8c5529bf0f4e",
  "0x89ece308f9d1f0131765212deca99697b112d61f9be9a5f1f3780a51335b3ff981747a0b2ca2179b96d2c0c9024e5224"
]
```

## Quick Test Commands

```bash
# Individual tests
curl localhost:8080/blockreward/7847950 | jq .
curl localhost:8080/syncduties/7847950 | jq .

# Test errors
curl localhost:8080/blockreward/invalid | jq .
curl localhost:8080/blockreward/99999999 | jq .
```

## Test with Postman

Import the Postman collection from `api_docs/Ethereum Validator API.postman_collection.json`.

## Implementation Details

### MEV Detection Heuristic
A block is classified as "MEV" if: The execution payload's `fee_recipient` differs from the proposer's withdrawal credentials address. Otherwise, it's classified as "VANILLA".

### Reward Calculation
Block reward is calculated as:
```
reward = (base_fee_per_gas × gas_used + total_priority_fees) ÷ 10^9
```
- Result is truncated to integer GWEI
- Handles both EIP-1559 and legacy transactions

### Sync Committee
- Sync committees rotate every 256 epochs (8192 slots)
- Returns up to 512 validator public keys
- Falls back to mock data if real data is unavailable (clearly marked in code)

## Error Format

All errors follow this JSON structure:
```json
{
  "error": "Description of the error",
  "code": 400
}
```

## Testing

```bash
go test -v ./...
```

## Performance Considerations

- HTTP client timeout: 30 seconds
- Graceful shutdown with 5-second timeout
- Efficient big number arithmetic for reward calculations
- Connection pooling via Go's default HTTP client

## Known Limitations

1. **Sync duties endpoint**: May return mock data if the beacon node doesn't expose sync committee data
2. **MEV detection**: Uses a simple heuristic based on fee recipient comparison
3. **Rate limiting**: No built-in rate limiting (rely on RPC provider's limits)

## Future Improvements

- Add caching layer for frequently accessed slots
- Implement rate limiting
- Add metrics and monitoring endpoints
- Support for multiple RPC endpoints with fallback
- WebSocket support for real-time updates

## License

MIT
