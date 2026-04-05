package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mark3labs/mcp-go/server"
)

// version is set at build time via -ldflags.
var version = "dev"

func main() {
	// Set up structured logging.
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Configuration from environment variables.
	apiURL := os.Getenv("SANDBOXAPI_URL")
	if apiURL == "" {
		apiURL = "https://api.sandboxapi.dev"
	}

	apiKey := os.Getenv("SANDBOXAPI_API_KEY")
	if apiKey == "" {
		slog.Error("SANDBOXAPI_API_KEY is required")
		os.Exit(1)
	}

	mcpAPIKey := os.Getenv("MCP_API_KEY")
	if mcpAPIKey == "" {
		slog.Warn("MCP_API_KEY not set — MCP endpoint has no authentication")
	}

	port := os.Getenv("MCP_PORT")
	if port == "" {
		port = "8081"
	}

	// Create HTTP client for SandboxAPI.
	client := NewSandboxClient(apiURL, apiKey)

	// Create MCP server with tools.
	mcpSrv := NewMCPServer(version, client)

	// Create Streamable HTTP transport.
	httpServer := server.NewStreamableHTTPServer(mcpSrv,
		server.WithStateLess(true),
		server.WithHTTPContextFunc(APIKeyAuthFunc(mcpAPIKey)),
	)

	// Wrap with auth middleware for proper 401 responses.
	mux := http.NewServeMux()
	mux.Handle("/mcp", AuthMiddleware(mcpAPIKey, httpServer))

	srv := &http.Server{
		Addr:           fmt.Sprintf(":%s", port),
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   5*time.Minute + 5*time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 13,
	}

	// Start server in goroutine.
	go func() {
		slog.Info("starting SandboxAPI MCP server",
			"port", port,
			"version", version,
			"api_url", apiURL,
			"auth", mcpAPIKey != "",
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("MCP server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down MCP server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("MCP server forced shutdown", "error", err)
	}

	slog.Info("MCP server stopped")
}
