package frontend

import (
	"embed"
	"log/slog"
	"net/http"
	"strings"
)

//go:embed dist/*
var Assets embed.FS

// GetHandler returns a handler for serving embedded frontend assets
func GetHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to serve the requested file from embedded assets
		filePath := "dist" + r.URL.Path
		if data, err := Assets.ReadFile(filePath); err == nil {
			// Set appropriate content type based on file extension
			setContentType(w, r.URL.Path)
			if _, err := w.Write(data); err != nil {
				slog.Error("Failed to write data", "error", err)
			}
			return
		}

		// If file doesn't exist, serve 404
		http.NotFound(w, r)
	})
}

// Config holds the frontend configuration
type Config struct {
	OIDCIssuer string
	ClientID   string
}

// Serve serves the frontend with SPA routing support for embedded assets
func Serve(w http.ResponseWriter, r *http.Request) {
	// Try to serve the requested file from embedded assets
	filePath := "dist" + r.URL.Path
	if data, err := Assets.ReadFile(filePath); err == nil {
		// Set appropriate content type based on file extension
		setContentType(w, r.URL.Path)
		if _, err := w.Write(data); err != nil {
			slog.Error("Failed to write data", "error", err)
		}
		return
	}

	// If file doesn't exist, serve index.html for SPA routing
	if data, err := Assets.ReadFile("dist/index.html"); err == nil {
		w.Header().Set("Content-Type", "text/html")
		if _, err := w.Write(data); err != nil {
			slog.Error("Failed to write data", "error", err)
		}
		return
	}

	// Fallback to 404
	http.NotFound(w, r)
}

// ServeWithConfig serves the frontend with injected configuration
func ServeWithConfig(w http.ResponseWriter, r *http.Request, config Config) {
	// Try to serve the requested file from embedded assets
	filePath := "dist" + r.URL.Path
	if data, err := Assets.ReadFile(filePath); err == nil {
		// Set appropriate content type based on file extension
		setContentType(w, r.URL.Path)
		if _, err := w.Write(data); err != nil {
			slog.Error("Failed to write data", "error", err)
		}
		return
	}

	// If file doesn't exist, serve index.html for SPA routing with injected config
	if data, err := Assets.ReadFile("dist/index.html"); err == nil {
		// Inject configuration into the HTML
		html := string(data)

		// Create a script tag with the configuration
		configScript := `<script>
window.__APP_CONFIG__ = {
  oidcIssuer: '` + config.OIDCIssuer + `',
  clientId: '` + config.ClientID + `'
};
</script>`

		// Inject the script before the closing </head> tag
		html = strings.Replace(html, "</head>", configScript+"</head>", 1)

		w.Header().Set("Content-Type", "text/html")
		if _, err := w.Write([]byte(html)); err != nil {
			slog.Error("Failed to write data", "error", err)
		}
		return
	}

	// Fallback to 404
	http.NotFound(w, r)
}

// setContentType sets the appropriate Content-Type header based on file extension
func setContentType(w http.ResponseWriter, path string) {
	switch {
	case path == "/" || path == "":
		w.Header().Set("Content-Type", "text/html")
	case path[len(path)-5:] == ".html":
		w.Header().Set("Content-Type", "text/html")
	case path[len(path)-4:] == ".css":
		w.Header().Set("Content-Type", "text/css")
	case path[len(path)-3:] == ".js":
		w.Header().Set("Content-Type", "application/javascript")
	case path[len(path)-4:] == ".svg":
		w.Header().Set("Content-Type", "image/svg+xml")
	case path[len(path)-4:] == ".png":
		w.Header().Set("Content-Type", "image/png")
	case path[len(path)-4:] == ".jpg" || path[len(path)-5:] == ".jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
	case path[len(path)-4:] == ".gif":
		w.Header().Set("Content-Type", "image/gif")
	case path[len(path)-4:] == ".ico":
		w.Header().Set("Content-Type", "image/x-icon")
	default:
		w.Header().Set("Content-Type", "text/plain")
	}
}
