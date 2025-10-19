# EQ Character Generation Service

A ConnectRPC service for generating EverQuest characters from JSON data dumps, with OAuth2 authentication and inventory management.

## Features

- **OAuth2 Authentication**: Secure token-based authentication using OIDC
- **Character Management**: Create, read, list, and delete characters
- **Inventory Support**: Full inventory management with items and augments
- **Database Integration**: MySQL database with proper schema and relationships
- **ConnectRPC**: Modern HTTP-based API using Protocol Buffers with ConnectRPC

## API Endpoints

All endpoints require OAuth2 Bearer token in the Authorization header.

### Character Management

- `CreateCharacter` - Create a new character from JSON data
- `GetCharacter` - Retrieve a character by ID
- `ListCharacters` - List user's characters with pagination
- `DeleteCharacter` - Remove a character and its inventory

## Setup

### Prerequisites

- Go 1.21+
- MySQL database
- OAuth2/OIDC provider (e.g., Auth0, Google, etc.)

### Database Setup

1. Create a MySQL database for your EQ emulator
2. Run the schema from `database/schema.sql`:

```sql
mysql -u your_username -p your_database < database/schema.sql
```

### Configuration

Set the following environment variables:

```bash
# OAuth2 Configuration
export OAUTH2_CLIENT_ID="eqtestcopy-spa"
export OAUTH2_CLIENT_SECRET="your-client-secret"
export OIDC_ISSUER="https://your-oauth-provider.com"

# Database Configuration
export DB_HOST="localhost"
export DB_NAME="eq_emulator"
export DB_USER="root"
export DB_PASSWORD="your-password"

# Server Configuration
export LISTEN_ADDR=":8080"
```

### Running the Service

1. Install dependencies:
```bash
go mod tidy
```

2. Generate protobuf code:
```bash
protoc --go_out=. --connect-go_out=. --go_opt=paths=source_relative --connect-go_opt=paths=source_relative proto/eqtestcopy/eqtestcopy.proto
```

3. Run the service:
```bash
go run cmd/eqtestcopy/main.go
```

## Usage

### Character Creation

The service accepts JSON character data in the following format:

```json
{
  "character_data": {
    "name": "TestCharacter",
    "race": 1,
    "class": 1,
    "level": 50,
    "zone_id": 1,
    "x": 0.0,
    "y": 0.0,
    "z": 0.0,
    "stats": {
      "hp": 1000,
      "mana": 500,
      "str": 100,
      "sta": 100,
      "agi": 100,
      "dex": 100,
      "wis": 100,
      "int": 100,
      "cha": 100
    },
    "inventory": [
      {
        "slot_id": 0,
        "item_id": 12345,
        "charges": 1,
        "augments": [67890, 11111]
      }
    ]
  },
  "validate": true
}
```

### Authentication

All requests must include a valid OAuth2 Bearer token:

```bash
Authorization: Bearer your-jwt-token
```

## Development

### Project Structure

```
├── cmd/eqtestcopy/          # Main application
├── pkg/internal/
│   ├── eqdb/                # Database layer
│   └── eqtestcopy/          # Server implementation
├── proto/eqtestcopy/        # Protocol buffer definitions
└── database/                # Database schema
```

### Adding New Features

1. Update the proto file with new RPCs and messages
2. Regenerate protobuf code
3. Implement handlers in `pkg/internal/eqtestcopy/handlers.go`
4. Add database queries in `pkg/internal/eqdb/queries.go`

## License

MIT License
