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

type OpenAIClient struct {
	apiKey string
	model  string
	url    string
	client *http.Client
}

func NewOpenAIClient(apiKey, model string) *OpenAIClient {
	if model == "" {
		model = "gpt-4o-mini"
	}
	return &OpenAIClient{
		apiKey: apiKey,
		model:  model,
		url:    "https://api.openai.com/v1/chat/completions",
		client: &http.Client{},
	}
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

type streamChoice struct {
	Delta struct {
		Content string `json:"content"`
	} `json:"delta"`
}

type streamChunk struct {
	Choices []streamChoice `json:"choices"`
}

func (c *OpenAIClient) Complete(ctx context.Context, systemPrompt string, messages []domain.Message, contextChunks []string) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		msgs := buildMessages(systemPrompt, messages, contextChunks)

		body, err := json.Marshal(chatRequest{
			Model:    c.model,
			Messages: msgs,
			Stream:   true,
		})
		if err != nil {
			yield("", fmt.Errorf("marshal request: %w", err))
			return
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(body))
		if err != nil {
			yield("", fmt.Errorf("create request: %w", err))
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+c.apiKey)

		resp, err := c.client.Do(req)
		if err != nil {
			yield("", fmt.Errorf("request: %w", err))
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			var errBody bytes.Buffer
			errBody.ReadFrom(resp.Body)
			yield("", fmt.Errorf("openai API %d: %s", resp.StatusCode, errBody.String()))
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				return
			}

			var chunk streamChunk
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				continue
			}
			if len(chunk.Choices) == 0 {
				continue
			}

			content := chunk.Choices[0].Delta.Content
			if content == "" {
				continue
			}
			if !yield(content, nil) {
				return
			}
		}
		if err := scanner.Err(); err != nil {
			yield("", fmt.Errorf("read stream: %w", err))
		}
	}
}

func buildMessages(systemPrompt string, history []domain.Message, contextChunks []string) []chatMessage {
	var msgs []chatMessage

	system := systemPrompt
	if len(contextChunks) > 0 {
		system += "\n\n## Relevant documents\n\n"
		for i, chunk := range contextChunks {
			system += fmt.Sprintf("--- Source %d ---\n%s\n\n", i+1, chunk)
		}
	}
	msgs = append(msgs, chatMessage{Role: "system", Content: system})

	for _, m := range history {
		if m.Role == domain.MessageRoleSystem {
			continue
		}
		msgs = append(msgs, chatMessage{Role: string(m.Role), Content: m.Content})
	}

	return msgs
}
