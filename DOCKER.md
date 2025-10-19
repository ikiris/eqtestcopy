# Docker Deployment Guide

This guide explains how to build and deploy the eqtestcopy application using Docker.

## Files Created

- `Dockerfile` - 2-stage build configuration
- `.dockerignore` - Optimizes build context
- `docker-compose.yml` - Easy deployment with environment variables
- `env.template` - Environment configuration template

## Building the Image

### Simple Build
```bash
docker build -t eqtestcopy .
```

### Build with Tag
```bash
docker build -t eqtestcopy:latest .
```

## Running the Container

### Using Docker Run
```bash
# Basic run (will fail without certificates)
docker run -p 3000:3000 eqtestcopy

# With environment variables
docker run -p 3000:3000 \
  -e DB_PASSWORD=your_password \
  -e OIDC_ISSUER=https://your-oidc-issuer \
  -e OAUTH2_CLIENT_ID=your-client-id \
  -v ./certs:/app/certs:ro \
  eqtestcopy
```

### Using Docker Compose
```bash
# Copy and edit environment file
cp env.template .env
# Edit .env with your configuration

# Create certificates directory
mkdir -p certs
# Copy your certificates to ./certs/

# Run with docker-compose
docker-compose up -d
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_USERNAME` | `quarm` | Database username |
| `DB_PASSWORD` | `quarm` | Database password |
| `DB_HOST` | `xev.teraptra.net` | Database host |
| `DB_PORT` | `3306` | Database port |
| `DB_NAME` | `quarm` | Database name |
| `OIDC_ISSUER` | `https://localhost:8443` | OIDC issuer URL |
| `OAUTH2_CLIENT_ID` | `eqtestcopy-spa` | OAuth2 client ID |

## Certificate Files

The application requires the following certificate files to be mounted in `/app/certs/`:

- `xev-teraptra-cert.pem` - TLS certificate
- `xev-teraptra-key.pem` - TLS private key  
- `teraptra-ca-cert.pem` - CA certificate

## Health Check

The container includes a health check that verifies the application is responding on port 3000.

## Security Features

- Runs as non-root user (appuser:appgroup)
- Minimal Alpine Linux base image
- Static Go binary with no dependencies
- Frontend assets embedded in binary

## Troubleshooting

### Frontend Build Issues
If you encounter dependency conflicts during the frontend build, the Dockerfile uses `--legacy-peer-deps` to resolve them.

### Certificate Issues
Ensure certificate files are properly mounted and have correct permissions. The container expects certificates in `/app/certs/`.

### Database Connection Issues
Verify database connectivity and credentials. The application will exit if it cannot connect to the database.

## Production Considerations

1. Use secrets management for sensitive environment variables
2. Use a reverse proxy (nginx/traefik) for SSL termination
3. Set up proper logging and monitoring
4. Use Docker secrets for certificate files in production
5. Consider using a container orchestration platform (Kubernetes, Docker Swarm)
