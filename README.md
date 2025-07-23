# Kafka Topic Reader

A Go service that provides HTTP access to read messages from Kafka topics. The service exposes a REST API to consume Kafka messages with pagination and filtering support.

## Features

- **HTTP API**: Read Kafka messages via REST endpoints
- **Binary Filtering**: Filter messages by binary pattern matching
- **Pagination**: Support for offset-based pagination with configurable limits
- **Monitoring**: Prometheus metrics and health check endpoints
- **Error Reporting**: Integration with Sentry for error tracking

## Quick Start

### Running the Service

```bash
# Run with command-line arguments
go run main.go -- \
  --listen=":8080" \
  --kafka-brokers="localhost:9092" \
  --sentry-dsn="your-sentry-dsn-here"

# Run with environment variables (alternative)
export KAFKA_BROKERS="localhost:9092"
export LISTEN=":8080"
export SENTRY_DSN="your-sentry-dsn-here"
go run main.go

# For local development, you can use a dummy Sentry DSN
go run main.go -- \
  --listen=":8080" \
  --kafka-brokers="localhost:9092" \
  --sentry-dsn="https://dummy@dummy.ingest.sentry.io/dummy"
```

### Docker

```bash
# Build Docker image
make build

# Run with Docker
docker run -p 8080:8080 -e KAFKA_BROKERS="localhost:9092" kafka-topic-reader
```

## API Reference

### Read Messages

```
GET /read
```

Read Kafka messages with query parameters:

**Parameters:**
- `topic` (required) - Kafka topic name
- `partition` (required) - Kafka partition number  
- `offset` (required) - Starting offset (supports negative values for relative positioning)
- `limit` (optional, default: 100) - Maximum number of records to return
- `filter` (optional, max: 1024 bytes) - Binary substring filter for raw message values (exact byte matching, case-sensitive)

**Example:**
```bash
# Read 10 messages from topic "events", partition 0, starting at offset 100
curl "http://localhost:8080/read?topic=events&partition=0&offset=100&limit=10"

# Filter messages containing "error"
curl "http://localhost:8080/read?topic=logs&partition=0&offset=0&filter=error"

# Use negative offset to read from end
curl "http://localhost:8080/read?topic=events&partition=0&offset=-10&limit=10"
```

**Response:**
```json
{
  "records": [
    {
      "key": "message-key",
      "value": {"data": "message content"},
      "offset": 100,
      "partition": 0,
      "topic": "events"
    }
  ],
  "nextOffset": 101
}
```

### Health Checks

- `GET /healthz` - Health check endpoint
- `GET /readiness` - Readiness check endpoint  
- `GET /metrics` - Prometheus metrics

### Log Level Management

- `POST /setloglevel/{level}` - Dynamic log level adjustment

## Configuration

The application supports both command-line arguments and environment variables:

### Required Parameters
- `--kafka-brokers` / `KAFKA_BROKERS` - Comma-separated list of Kafka broker addresses
- `--listen` / `LISTEN` - HTTP server listen address (e.g., ":8080")
- `--sentry-dsn` / `SENTRY_DSN` - Sentry error reporting DSN

### Optional Parameters  
- `--sentry-proxy` / `SENTRY_PROXY` - Sentry proxy URL

**Note for Development**: While Sentry DSN is marked as required, you can use a dummy DSN for local development.

**Note**: Command-line arguments take precedence over environment variables.

## Binary Filtering

The service supports binary pattern matching on raw Kafka message values:

- **Case-sensitive**: Exact byte matching without case conversion
- **Binary safe**: Works with any binary data, not just text
- **Efficient**: Filtering happens before message conversion
- **Size limit**: Filter parameter limited to 1024 bytes for security

**Examples:**
```bash
# Filter JSON messages containing specific field
curl "http://localhost:8080/read?topic=api-logs&partition=0&offset=0&filter=user_id"

# Filter binary data (URL-encoded)
curl "http://localhost:8080/read?topic=binary-data&partition=0&offset=0&filter=%00%01%FF"
```

## Development

### Building and Testing

```bash
# Run full precommit pipeline (required before commits)
make precommit

# Individual commands
make ensure        # Tidy and verify go modules
make format        # Format code with goimports-reviser
make generate      # Generate code (mocks, etc.)
make test          # Run tests with race detection and coverage
make check         # Run vet, errcheck, and vulnerability checks

# Run all tests
go test -mod=mod ./...
```

### Docker Operations

```bash
make build         # Build Docker image
make upload        # Push Docker image to registry
make clean         # Remove Docker image
```

## Architecture

The service is built using:

- **Kafka Client**: IBM Sarama library for Kafka operations
- **HTTP Router**: Gorilla Mux for request routing
- **Testing**: Ginkgo v2 with Gomega for BDD-style testing
- **Monitoring**: Prometheus metrics integration
- **Error Handling**: Context-aware error handling with Sentry integration

## License

This project is licensed under the BSD-style license. See the LICENSE file for details.