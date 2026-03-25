package domain

import (
	"context"
	"iter"
)

type LlmService interface {
	Complete(ctx context.Context, systemPrompt string, messages []Message, contextChunks []string) iter.Seq2[string, error]
}

type EmbeddingService interface {
	Embed(ctx context.Context, texts []string) ([][]float32, error)
}

type VectorStore interface {
	Add(id string, orgID string, embedding []float32)
	Search(query []float32, orgID string, topK int) []VectorResult
	Remove(ids []string)
	Save(path string) error
	Load(path string) error
}

type DocumentParser interface {
	Parse(content []byte, mimeType string) (string, error)
}

type DocumentStore interface {
	CreateOrganization(ctx context.Context, org Organization) error
	GetOrganization(ctx context.Context, id string) (Organization, error)

	UpsertProfile(ctx context.Context, profile OrganizationProfile) error
	GetProfile(ctx context.Context, orgID string) (OrganizationProfile, error)

	CreateUser(ctx context.Context, user User) error
	GetUser(ctx context.Context, id string) (User, error)
	GetUserByToken(ctx context.Context, tokenHash string) (User, error)

	CreateDocument(ctx context.Context, doc Document) error
	UpdateDocumentStatus(ctx context.Context, id string, status DocumentStatus, errMsg string) error
	ListDocuments(ctx context.Context, orgID string) ([]Document, error)
	DeleteDocument(ctx context.Context, id string) error

	InsertChunks(ctx context.Context, chunks []Chunk) error
	GetChunksByIDs(ctx context.Context, ids []string) ([]Chunk, error)
	DeleteChunksByDocument(ctx context.Context, documentID string) ([]string, error)

	CreateConversation(ctx context.Context, conv Conversation) error
	ListConversations(ctx context.Context, userID string) ([]Conversation, error)

	CreateMessage(ctx context.Context, msg Message) error
	ListMessages(ctx context.Context, conversationID string) ([]Message, error)
}
