package eqtestcopy

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"

	"connectrpc.com/connect"
	"github.com/coreos/go-oidc/v3/oidc"

	"github.com/ikiris/eqmaclib/eqdb"
	pbconnect "github.com/ikiris/eqtestcopy/proto/eqtestcopy/eqtestcopyconnect"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

type eqDBClient interface {
	GetCharacter(ctx context.Context, accountID int32, characterID int32) (*eqdb.Character, eqdb.Inventory, error)
	ListCharacters(ctx context.Context, req eqdb.ListCharactersRequest) ([]eqdb.Character, error)
	UpdateInventory(ctx context.Context, characterID int32, accountID int32, inventory eqdb.Inventory) error
	GetItemNames(ctx context.Context, itemIDs []int32) (map[int32]string, error)
}

type server struct {
	pbconnect.UnimplementedEqTestCopyServiceHandler
	db       eqDBClient
	verifier *oidc.IDTokenVerifier
}

// New creates a new server instance
func New(db *sql.DB, verifier *oidc.IDTokenVerifier) (*server, error) {
	eqDB, err := eqdb.New(db)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return &server{
		db:       eqDB,
		verifier: verifier,
	}, nil
}

// OAuth2Interceptor validates OAuth2 tokens in HTTP headers
func (s *server) OAuth2Interceptor(next connect.UnaryFunc) connect.UnaryFunc {
	return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		// Get authorization header from HTTP request
		authHeader := req.Header().Get("Authorization")
		if authHeader == "" {
			return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("missing authorization header"))
		}

		// Extract token from "Bearer <token>" format
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("invalid authorization header format"))
		}

		// Verify the token
		idToken, err := s.verifier.Verify(ctx, token)
		if err != nil {
			slog.Warn("Token verification failed", "error", err)
			return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("invalid token"))
		}

		// Extract user ID from token claims
		var claims struct {
			Subject string `json:"sub"`
		}
		if err := idToken.Claims(&claims); err != nil {
			return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("invalid token claims"))
		}

		// Add user ID to context for handlers to use
		ctx = context.WithValue(ctx, contextKey("account_id"), claims.Subject)

		return next(ctx, req)
	})
}
