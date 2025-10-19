# OAuth Client Validation System

This document describes the client validation system implemented for the OAuth2/OIDC provider.

## Configuration

Clients are configured in `config/clients.yaml`. Here's the structure:

```yaml
clients:
  - id: "eqtestcopy"
    secret: "eqtestcopy-secret-key"
    name: "EQ Test Copy Client"
    redirect_uris:
      - "http://localhost:3000/callback"
      - "http://localhost:8080/callback"
      - "https://localhost:3000/callback"
      - "https://localhost:8080/callback"
    grant_types:
      - "authorization_code"
      - "refresh_token"
    response_types:
      - "code"
    scopes:
      - "openid"
      - "profile"
      - "email"
    require_https: false  # Set to true in production
```

## Security Validations Implemented

### 1. Client ID Validation
- Validates that the `client_id` is registered in the configuration
- Prevents unauthorized clients from accessing the OAuth endpoints

### 2. Client Type Authentication
- **Confidential Clients**: Server-side applications that can securely store secrets
  - Must provide `client_secret` during token exchange
  - Used for traditional web applications, backend services
- **Public Clients**: SPAs, mobile apps that cannot securely store secrets
  - Do NOT provide `client_secret` (rejected if provided)
  - Should use PKCE for additional security

### 3. Redirect URI Validation
- Ensures the `redirect_uri` matches one of the client's registered URIs
- Prevents open redirect attacks
- Normalizes URIs for comparison (removes trailing slashes, converts to lowercase)

### 4. HTTPS Enforcement
- Can be configured per client to require HTTPS redirect URIs
- Set `require_https: true` in production environments

### 5. Grant Type Validation
- Ensures clients only use allowed grant types
- Prevents unauthorized grant type usage

### 6. Response Type Validation
- Validates response types during authorization requests
- Ensures compliance with OAuth2/OIDC specifications

## Usage

The validation functions are automatically used in the OAuth endpoints:

- `/auth` - Authorization endpoint validates client_id, response_type, and redirect_uri
- `/token` - Token endpoint validates client authentication based on client type:
  - Confidential clients: validates client_id and client_secret
  - Public clients: validates client_id only (rejects client_secret)
- `/token` (refresh) - Refresh token endpoint uses same client authentication logic

## OAuth2 Flow Examples

### Confidential Client (Server-side App)
```bash
# Authorization request (user's browser)
GET /auth?client_id=eqtestcopy&response_type=code&redirect_uri=https://myapp.com/callback

# Token exchange (server-to-server)
POST /token
{
  "grant_type": "authorization_code",
  "code": "auth_code_123",
  "client_id": "eqtestcopy",
  "client_secret": "eqtestcopy-secret-key",
  "redirect_uri": "https://myapp.com/callback"
}
```

### Public Client (SPA/Mobile)
```bash
# Authorization request (user's browser)
GET /auth?client_id=eqtestcopy-spa&response_type=code&redirect_uri=https://myapp.com/callback

# Token exchange (client app)
POST /token
{
  "grant_type": "authorization_code", 
  "code": "auth_code_123",
  "client_id": "eqtestcopy-spa",
  "redirect_uri": "https://myapp.com/callback"
  # Note: NO client_secret for public clients
}
```

## Configuration File Location

By default, the client configuration is loaded from `config/clients.yaml`. You can specify a different location using:

- Command line flag: `-config /path/to/config.yaml`
- Environment variable: `CLIENT_CONFIG=/path/to/config.yaml`

## Security Best Practices

1. **Use HTTPS in Production**: Set `require_https: true` for production clients
2. **Strong Client Secrets**: Use cryptographically secure random secrets
3. **Minimal Redirect URIs**: Only register the exact URIs needed
4. **Regular Secret Rotation**: Change client secrets periodically
5. **Monitor Access**: Log and monitor OAuth requests for anomalies

## Example Client Configuration for Production

```yaml
clients:
  - id: "production-app"
    secret: "your-very-secure-random-secret-here"
    name: "Production Application"
    redirect_uris:
      - "https://yourdomain.com/auth/callback"
    grant_types:
      - "authorization_code"
      - "refresh_token"
    response_types:
      - "code"
    scopes:
      - "openid"
      - "profile"
      - "email"
    require_https: true
```
