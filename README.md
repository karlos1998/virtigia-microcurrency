# Virtigia Microcurrency Service

A lightweight microservice for handling in-game microcurrency transactions.

## Features

- Add currency to wallets
- Remove currency from wallets
- View wallet balance
- View transaction history with pagination
- Embedded database with wallet ID indexing
- Bearer token authentication
- Docker support for easy deployment

## Getting Started

### Prerequisites

- Go 1.20 or higher
- Docker and Docker Compose (for production deployment)

### Environment Variables

Copy the example environment file and adjust as needed:

```bash
cp .env.example .env
```

Available configuration options:

- `PORT`: The port on which the server will listen (default: 8880)
- `DATA_DIR`: The directory where the database files will be stored (default: ./data)
- `API_TOKEN`: The bearer token used for authentication

### Running Locally

```bash
go run main.go
```

### Running with Docker Compose

```bash
docker-compose up -d
```

## API Documentation

### Swagger UI

The API documentation is available through Swagger UI at the root URL of the service:

```
http://localhost:8880/
```

This will redirect you to the Swagger UI interface where you can explore and test all available endpoints.

### Authentication

All API endpoints require authentication using a bearer token. Include the token in the `Authorization` header:

```
Authorization: Bearer your-token-here
```

### Add Currency to Wallet

**Endpoint**: `POST /api/v1/wallets/{wallet_id}/add`

**Path Parameters**:
- `wallet_id`: The ID of the wallet

**Request Body**:
```json
{
  "amount": 100.0,
  "description": "Game reward",
  "additional_data": {
    "game_id": "game456",
    "level": 5
  }
}
```

**Response**:
```json
{
  "transaction": {
    "id": "20230101120000",
    "wallet_id": "wallet123",
    "amount": 100.0,
    "description": "Game reward",
    "additional_data": {
      "game_id": "game456",
      "level": 5
    },
    "timestamp": "2023-01-01T12:00:00Z"
  },
  "wallet": {
    "wallet_id": "wallet123",
    "balance": 100.0
  }
}
```

### Remove Currency from Wallet

**Endpoint**: `POST /api/v1/wallets/{wallet_id}/remove`

**Path Parameters**:
- `wallet_id`: The ID of the wallet

**Request Body**:
```json
{
  "amount": 50.0,
  "description": "Item purchase",
  "additional_data": {
    "item_id": "item789"
  }
}
```

**Response**:
```json
{
  "transaction": {
    "id": "20230101120100",
    "wallet_id": "wallet123",
    "amount": -50.0,
    "description": "Item purchase",
    "additional_data": {
      "item_id": "item789"
    },
    "timestamp": "2023-01-01T12:01:00Z"
  },
  "wallet": {
    "wallet_id": "wallet123",
    "balance": 50.0
  }
}
```

### Get Wallet Balance

**Endpoint**: `GET /api/v1/wallets/{wallet_id}/balance`

**Path Parameters**:
- `wallet_id`: The ID of the wallet

**Response**:
```json
{
  "wallet_id": "wallet123",
  "balance": 50.0
}
```

### Get Transaction History

**Endpoint**: `GET /api/v1/wallets/{wallet_id}/transactions?limit=50&offset=0`

**Path Parameters**:
- `wallet_id`: The ID of the wallet

**Query Parameters**:
- `limit`: Maximum number of transactions to return (default: 50)
- `offset`: Number of transactions to skip (default: 0)

**Response**:
```json
{
  "transactions": [
    {
      "id": "20230101120100",
      "wallet_id": "wallet123",
      "amount": -50.0,
      "description": "Item purchase",
      "additional_data": {
        "item_id": "item789"
      },
      "timestamp": "2023-01-01T12:01:00Z"
    },
    {
      "id": "20230101120000",
      "wallet_id": "wallet123",
      "amount": 100.0,
      "description": "Game reward",
      "additional_data": {
        "game_id": "game456",
        "level": 5
      },
      "timestamp": "2023-01-01T12:00:00Z"
    }
  ],
  "wallet": {
    "wallet_id": "wallet123",
    "balance": 50.0
  },
  "pagination": {
    "limit": 50,
    "offset": 0,
    "count": 2
  }
}
```

## Error Handling

All errors are returned in a consistent format:

```json
{
  "error": "Error message"
}
```

Common error responses:
- `400 Bad Request`: Invalid request parameters
- `401 Unauthorized`: Missing or invalid authentication token
- `500 Internal Server Error`: Server-side error

## Running Tests

```bash
go test ./...
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.