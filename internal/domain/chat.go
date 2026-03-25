package domain

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"iter"
	"time"

	"github.com/lex-rio/omni-knowledge-base-assistant/internal/domain/id"
)

const defaultSystemPrompt = `You are Omni, a knowledge assistant for businesses. You run on a local device owned by the user — their data never leaves their network.

Be concise, helpful, and friendly. Answer in the same language the user writes in.`

const onboardingPrompt = `You are Omni, a knowledge assistant for businesses. You run on a local device owned by the user — their data never leaves their network.

Your current goal: learn about the user and their business through natural conversation. This is the first interaction.

Collect the following through friendly dialogue (one question at a time, do not overwhelm):
1. User's name
2. What their company/business does
3. What types of documents they work with (contracts, invoices, manuals, etc.)
4. What questions they typically need answered

Be concise and friendly. Ask one question at a time. Answer in the same language the user writes in.

When you have enough context, summarize what you learned and suggest the user upload their first documents.`

type ChatDeps struct {
	Store DocumentStore
	LLM   LlmService
}

type ChatInput struct {
	UserID         string
	ConversationID string
	Message        string
}

func HandleChat(ctx context.Context, input ChatInput, deps ChatDeps) (conversationID string, stream iter.Seq2[string, error]) {
	convID := input.ConversationID

	if convID == "" {
		convID = id.New()
		_ = deps.Store.CreateConversation(ctx, Conversation{
			ID:        convID,
			UserID:    input.UserID,
			CreatedAt: time.Now(),
		})
	}

	_ = deps.Store.CreateMessage(ctx, Message{
		ID:             id.New(),
		ConversationID: convID,
		Role:           MessageRoleUser,
		Content:        input.Message,
		CreatedAt:      time.Now(),
	})

	history, _ := deps.Store.ListMessages(ctx, convID)

	user, _ := deps.Store.GetUser(ctx, input.UserID)
	systemPrompt := pickSystemPrompt(ctx, user.OrgID, deps.Store)

	llmStream := deps.LLM.Complete(ctx, systemPrompt, history, nil)

	return convID, collectAndSave(ctx, convID, llmStream, deps.Store)
}

func pickSystemPrompt(ctx context.Context, orgID string, store DocumentStore) string {
	profile, err := store.GetProfile(ctx, orgID)
	if errors.Is(err, sql.ErrNoRows) || !profile.OnboardingCompleted {
		return onboardingPrompt
	}

	prompt := defaultSystemPrompt
	if profile.BusinessDescription != "" {
		prompt += fmt.Sprintf("\n\nOrganization context: %s", profile.BusinessDescription)
	}
	if profile.DocumentTypes != "" {
		prompt += fmt.Sprintf("\nDocument types they work with: %s", profile.DocumentTypes)
	}
	if profile.ContactName != "" {
		prompt += fmt.Sprintf("\nUser's name: %s", profile.ContactName)
	}
	return prompt
}

func collectAndSave(ctx context.Context, convID string, stream iter.Seq2[string, error], store DocumentStore) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		var full string
		for chunk, err := range stream {
			if err != nil {
				yield("", err)
				return
			}
			full += chunk
			if !yield(chunk, nil) {
				return
			}
		}

		_ = store.CreateMessage(ctx, Message{
			ID:             id.New(),
			ConversationID: convID,
			Role:           MessageRoleAssistant,
			Content:        full,
			CreatedAt:      time.Now(),
		})
	}
}
