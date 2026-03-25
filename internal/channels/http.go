package channels

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"runtime"

	"github.com/lex-rio/omni-knowledge-base-assistant/internal/domain"
)

type Deps struct {
	Store   domain.DocumentStore
	Vectors domain.VectorStore
	LLM     domain.LlmService
	WebFS   fs.FS
}

func NewHTTPServer(port int, deps Deps) *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", handleHealth)
	mux.HandleFunc("POST /api/chat", handleChat(deps))

	if deps.WebFS != nil {
		mux.Handle("/", http.FileServer(http.FS(deps.WebFS)))
	}

	return &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}
}

type chatRequest struct {
	Message        string `json:"message"`
	ConversationID string `json:"conversation_id"`
}

func handleChat(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}
		if req.Message == "" {
			http.Error(w, `{"error":"message is required"}`, http.StatusBadRequest)
			return
		}

		// TODO: replace with real auth; for now use a default user
		userID := "default-user"

		input := domain.ChatInput{
			UserID:         userID,
			ConversationID: req.ConversationID,
			Message:        req.Message,
		}
		chatDeps := domain.ChatDeps{
			Store: deps.Store,
			LLM:   deps.LLM,
		}

		convID, stream := domain.HandleChat(r.Context(), input, chatDeps)

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Conversation-ID", convID)

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming not supported", http.StatusInternalServerError)
			return
		}

		for chunk, err := range stream {
			if err != nil {
				slog.Error("chat stream", "error", err)
				fmt.Fprintf(w, "data: [ERROR] %s\n\n", err.Error())
				flusher.Flush()
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", chunk)
			flusher.Flush()
		}

		fmt.Fprint(w, "data: [DONE]\n\n")
		flusher.Flush()
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	resp := map[string]any{
		"status": "ok",
		"arch":   runtime.GOARCH,
		"os":     runtime.GOOS,
		"memory": map[string]any{
			"alloc_mb":       memStats.Alloc / 1024 / 1024,
			"sys_mb":         memStats.Sys / 1024 / 1024,
			"num_goroutines": runtime.NumGoroutine(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("health encode", "error", err)
	}
}
