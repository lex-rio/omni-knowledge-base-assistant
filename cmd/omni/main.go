package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lex-rio/omni-knowledge-base-assistant/assets"
	"github.com/lex-rio/omni-knowledge-base-assistant/internal/adapters/llm"
	"github.com/lex-rio/omni-knowledge-base-assistant/internal/adapters/storage"
	"github.com/lex-rio/omni-knowledge-base-assistant/internal/channels"
	"github.com/lex-rio/omni-knowledge-base-assistant/internal/config"
	"github.com/lex-rio/omni-knowledge-base-assistant/internal/domain"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("config", "error", err)
		os.Exit(1)
	}

	if err := os.MkdirAll(cfg.DocsDir(), 0o750); err != nil {
		slog.Error("create data dirs", "error", err)
		os.Exit(1)
	}

	store, err := storage.NewSQLiteStore(cfg.DBPath(), assets.MigrationSQL)
	if err != nil {
		slog.Error("sqlite", "error", err)
		os.Exit(1)
	}
	defer store.Close()

	ensureDefaultUser(store)

	vecStore := storage.NewMemoryVectorStore()
	if err := vecStore.Load(cfg.VecPath()); err != nil {
		slog.Warn("vector index load", "error", err)
	}

	llmClient, err := createLLM(cfg)
	if err != nil {
		slog.Error("llm", "error", err)
		os.Exit(1)
	}

	webContent, _ := fs.Sub(assets.WebFS, "web")

	deps := channels.Deps{
		Store:   store,
		Vectors: vecStore,
		LLM:     llmClient,
		WebFS:   webContent,
	}

	srv := channels.NewHTTPServer(cfg.Port, deps)

	go func() {
		slog.Info("starting server", "port", cfg.Port, "llm", cfg.LLMProvider)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := vecStore.Save(cfg.VecPath()); err != nil {
		slog.Error("vector index save", "error", err)
	}

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("shutdown", "error", err)
	}
}

func createLLM(cfg config.Config) (domain.LlmService, error) {
	switch cfg.LLMProvider {
	case "gemini":
		if cfg.GeminiKey == "" {
			return nil, fmt.Errorf("GEMINI_API_KEY is required when LLM_PROVIDER=gemini")
		}
		return llm.NewGeminiClient(cfg.GeminiKey, ""), nil
	case "openai":
		if cfg.OpenAIKey == "" {
			return nil, fmt.Errorf("OPENAI_API_KEY is required when LLM_PROVIDER=openai")
		}
		return llm.NewOpenAIClient(cfg.OpenAIKey, ""), nil
	default:
		return nil, fmt.Errorf("unknown LLM_PROVIDER: %s (supported: gemini, openai)", cfg.LLMProvider)
	}
}

func ensureDefaultUser(store domain.DocumentStore) {
	ctx := context.Background()
	_, err := store.GetOrganization(ctx, "default-org")
	if err != nil {
		_ = store.CreateOrganization(ctx, domain.Organization{
			ID:        "default-org",
			Name:      "Default",
			CreatedAt: time.Now(),
		})
		_ = store.CreateUser(ctx, domain.User{
			ID:        "default-user",
			OrgID:     "default-org",
			Name:      "User",
			Role:      "admin",
			CreatedAt: time.Now(),
		})
		slog.Info("created default organization and user")
	}
}
