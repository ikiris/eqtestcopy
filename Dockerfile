# Stage 1: Build frontend
FROM node:20-alpine AS frontend-builder

ENV GOOS=linux
ENV GOARCH=amd64

# Set working directory
WORKDIR /app/frontend

# Copy package files
COPY frontend/package*.json ./

# Install dependencies with cache mount
RUN --mount=type=cache,target=/root/.npm \
    npm ci --legacy-peer-deps

# Copy frontend source code
COPY frontend/ ./

# Build the frontend
RUN npm run build

# Stage 2: Build Go application
FROM golang:1.25-alpine AS go-builder

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies with cache mount
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy only necessary source files for Go build
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY proto/ ./proto/

# Copy frontend Go source and built assets for embedding
COPY frontend/build.go ./frontend/build.go
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist

# Build the Go application
RUN go build ./cmd/eqtestcopy

# Final stage: Runtime
FROM alpine:3.19

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=go-builder /app/eqtestcopy .

# Copy TLS certificate files (these should be provided via volume mounts in production)
# COPY *.pem ./

# Change ownership to non-root user
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 3000

# Run the application
CMD ["./eqtestcopy"]
