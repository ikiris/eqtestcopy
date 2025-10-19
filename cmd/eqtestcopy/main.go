package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"connectrpc.com/connect"
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/sync/errgroup"

	"github.com/ikiris/eqtestcopy/frontend"
	"github.com/ikiris/eqtestcopy/internal/eqtestcopy"
	pbconnect "github.com/ikiris/eqtestcopy/proto/eqtestcopy/eqtestcopyconnect"
)

func main() {
	ctx := context.Background()

	if err := doStuff(ctx); err != nil {
		slog.Error("Server failed", "error", err)
		os.Exit(1)
	}
}

func doStuff(ctx context.Context) error {

	dbUsername := flag.String("db-username", "quarm", "Database username")
	dbHost := flag.String("db-host", "xev.teraptra.net", "Database host")
	dbPort := flag.Uint("db-port", 3306, "Database port")
	dbDatabase := flag.String("db-name", "quarm", "Database name")
	addr := flag.String("addr", "0.0.0.0", "Address to listen on")
	port := flag.Uint("port", 3000, "Port to listen on")
	useTLS := flag.Bool("tls", true, "Use TLS")

	// TLS configuration flags
	tlsCertFile := flag.String("tls-cert", getEnv("CERT_FILE", "xev-teraptra-cert.pem"), "TLS certificate file path")
	tlsKeyFile := flag.String("tls-key", getEnv("KEY_FILE", "xev-teraptra-key.pem"), "TLS private key file path")
	tlsCAFile := flag.String("ca-cert", getEnv("CA_FILE", "teraptra-ca-cert.pem"), "TLS private key file path")

	// OIDC configuration flags
	oidcIssuer := flag.String("oidc-issuer", getEnv("OIDC_ISSUER", "https://localhost:8443"), "OIDC issuer URL")
	oidcClientID := flag.String("oidc-client-id", getEnv("OAUTH2_CLIENT_ID", "eqtestcopy-spa"), "OIDC client ID")

	flag.Parse()

	// Database configuration
	dbPassword := getEnv("DB_PASSWORD", "quarm")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		*dbUsername, dbPassword, *dbHost, *dbPort, *dbDatabase)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			slog.Error("failed to close database", "error", err)
		}
	}()

	// Initialize OIDC provider with custom HTTP client if CA file is provided
	var provider *oidc.Provider
	// Load CA certificate
	caData, err := os.ReadFile(*tlsCAFile)
	if err != nil {
		return fmt.Errorf("failed to read CA file %s: %w", *tlsCAFile, err)
	}

	caPool, err := x509.SystemCertPool()
	if err != nil {
		return fmt.Errorf("failed to get system cert pool: %w", err)
	}

	if caPool == nil {
		caPool = x509.NewCertPool()
	}

	if !caPool.AppendCertsFromPEM(caData) {
		return fmt.Errorf("failed to parse CA certificate from %s", *tlsCAFile)
	}

	// Create custom HTTP client with CA
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:    caPool,
				MinVersion: tls.VersionTLS12,
			},
		},
	}

	// Create OIDC provider with custom HTTP client
	provider, err = oidc.NewProvider(oidc.ClientContext(ctx, httpClient), *oidcIssuer)
	if err != nil {
		return fmt.Errorf("failed to initialize OIDC provider with CA: %w", err)
	}

	// Create OIDC verifier
	verifier := provider.Verifier(&oidc.Config{ClientID: *oidcClientID})

	// Initialize server
	server, err := eqtestcopy.New(db, verifier)
	if err != nil {
		return err
	}

	// Create ConnectRPC handler with OAuth2 interceptor
	interceptors := connect.WithInterceptors(connect.UnaryInterceptorFunc(server.OAuth2Interceptor))

	// Create service handler
	path, serviceHandler := pbconnect.NewEqTestCopyServiceHandler(server, interceptors)

	// Create HTTP server
	mux := http.NewServeMux()
	mux.Handle(path, serviceHandler)

	// Get frontend handler (embedded in production, filesystem in dev)
	frontendHandler := frontend.GetHandler()

	// Handle static assets
	mux.Handle("/assets/", frontendHandler)
	mux.Handle("/static/", frontendHandler)

	// For SPA routes, serve index.html to let React Router handle routing
	mux.HandleFunc("/", frontend.Serve)

	// Add logging middleware
	loggedMux := loggingMiddleware(mux)

	// Start listening
	slog.Info("Starting ConnectRPC server", "addr", addr)

	tlsConfig, err := setupTLSConfig(*tlsCAFile, *tlsCertFile, *tlsKeyFile)
	if err != nil {
		return fmt.Errorf("failed to setup TLS config: %w", err)
	}

	shutdownCtx, shutdownCtxCancel := signal.NotifyContext(ctx, os.Interrupt)

	if !*useTLS {
		tlsConfig = nil
	}

	// Start HTTP server
	httpServer := &http.Server{
		Addr:      *addr + ":" + strconv.Itoa(int(*port)),
		Handler:   loggedMux,
		TLSConfig: tlsConfig,
	}

	defer shutdownCtxCancel()

	errG, gCtx := errgroup.WithContext(ctx)
	errG.Go(func() error {
		slog.Info("Starting HTTP server", "addr", addr)

		if !*useTLS {
			return httpServer.ListenAndServe()
		}

		return httpServer.ListenAndServeTLS("", "")
	})

	errG.Go(func() error {
		select {
		case <-shutdownCtx.Done():
		case <-gCtx.Done():
		}

		sCtx, sCtxCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer sCtxCancel()

		if err := httpServer.Shutdown(sCtx); err != nil {
			return fmt.Errorf("failed to shutdown HTTP server: %w", err)
		}

		return nil
	})

	if err := errG.Wait(); err != nil {
		return fmt.Errorf("server failed: %w", err)
	}

	return nil
}

// setupTLSConfig creates a TLS configuration from certificate and key files
func setupTLSConfig(caFile, certFile, keyFile string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS certificate: %w", err)
	}

	ca, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS CA: %w", err)
	}

	caPool, err := x509.SystemCertPool()
	if err != nil {
		return nil, fmt.Errorf("failed to get system cert pool: %w", err)
	}

	caPool.AppendCertsFromPEM(ca)

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caPool,
		MinVersion:   tls.VersionTLS12, // Require TLS 1.2 or higher
	}, nil
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// loggingMiddleware logs all HTTP requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		wrapper := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Call the next handler
		next.ServeHTTP(wrapper, r)

		// Log the request
		slog.Info("HTTP Request",
			"method", r.Method,
			"url", r.URL.String(),
			"status", wrapper.statusCode,
			"duration", time.Since(start),
			"user_agent", r.UserAgent(),
			"remote_addr", r.RemoteAddr,
		)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
