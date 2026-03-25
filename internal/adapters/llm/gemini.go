package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"net/http"
	"strings"

	"github.com/lex-rio/omni-knowledge-base-assistant/internal/domain"
)

type GeminiClient struct {
	apiKey string
	model  string
	client *http.Client
}

func NewGeminiClient(apiKey, model string) *GeminiClient {
	if model == "" {
		model = "gemini-3-flash-preview"
	}
	return &GeminiClient{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{},
	}
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiRequest struct {
	SystemInstruction *geminiContent  `json:"system_instruction,omitempty"`
	Contents          []geminiContent `json:"contents"`
}

type geminiCandidate struct {
	Content struct {
		Parts []geminiPart `json:"parts"`
	} `json:"content"`
}

type geminiStreamChunk struct {
	Candidates []geminiCandidate `json:"candidates"`
}

func (c *GeminiClient) Complete(ctx context.Context, systemPrompt string, messages []domain.Message, contextChunks []string) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		system := systemPrompt
		if len(contextChunks) > 0 {
			system += "\n\n## Relevant documents\n\n"
			for i, chunk := range contextChunks {
				system += fmt.Sprintf("--- Source %d ---\n%s\n\n", i+1, chunk)
			}
		}

		reqBody := geminiRequest{
			SystemInstruction: &geminiContent{
				Parts: []geminiPart{{Text: system}},
			},
			Contents: toGeminiContents(messages),
		}

		body, err := json.Marshal(reqBody)
		if err != nil {
			yield("", fmt.Errorf("marshal request: %w", err))
			return
		}

		url := fmt.Sprintf(
			"https://generativelanguage.googleapis.com/v1beta/models/%s:streamGenerateContent?alt=sse",
			c.model,
		)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
		if err != nil {
			yield("", fmt.Errorf("create request: %w", err))
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("x-goog-api-key", c.apiKey)

		resp, err := c.client.Do(req)
		if err != nil {
			yield("", fmt.Errorf("request: %w", err))
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			var errBody bytes.Buffer
			errBody.ReadFrom(resp.Body)
			yield("", fmt.Errorf("gemini API %d: %s", resp.StatusCode, errBody.String()))
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			data := strings.TrimPrefix(line, "data: ")

			var chunk geminiStreamChunk
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				continue
			}
			if len(chunk.Candidates) == 0 || len(chunk.Candidates[0].Content.Parts) == 0 {
				continue
			}

			text := chunk.Candidates[0].Content.Parts[0].Text
			if text == "" {
				continue
			}
			if !yield(text, nil) {
				return
			}
		}
		if err := scanner.Err(); err != nil {
			yield("", fmt.Errorf("read stream: %w", err))
		}
	}
}

func toGeminiContents(messages []domain.Message) []geminiContent {
	var contents []geminiContent
	for _, m := range messages {
		if m.Role == domain.MessageRoleSystem {
			continue
		}
		role := "user"
		if m.Role == domain.MessageRoleAssistant {
			role = "model"
		}
		contents = append(contents, geminiContent{
			Role:  role,
			Parts: []geminiPart{{Text: m.Content}},
		})
	}
	return contents
}
