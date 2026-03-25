package domain

import "time"

type Organization struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type OrganizationProfile struct {
	OrgID               string `json:"org_id"`
	ContactName         string `json:"contact_name"`
	BusinessDescription string `json:"business_description"`
	DocumentTypes       string `json:"document_types"`
	Preferences         string `json:"preferences"`
	OnboardingCompleted bool   `json:"onboarding_completed"`
	RawAnswers          string `json:"raw_answers"`
	UpdatedAt          string `json:"updated_at"`
}

type User struct {
	ID            string    `json:"id"`
	OrgID         string    `json:"org_id"`
	Name          string    `json:"name"`
	Role          string    `json:"role"`
	AuthTokenHash string    `json:"-"`
	CreatedAt     time.Time `json:"created_at"`
}

type DocumentStatus string

const (
	DocumentStatusPending    DocumentStatus = "pending"
	DocumentStatusProcessing DocumentStatus = "processing"
	DocumentStatusReady      DocumentStatus = "ready"
	DocumentStatusError      DocumentStatus = "error"
)

type Document struct {
	ID           string         `json:"id"`
	OrgID        string         `json:"org_id"`
	Filename     string         `json:"filename"`
	MimeType     string         `json:"mime_type"`
	SizeBytes    int64          `json:"size_bytes"`
	Status       DocumentStatus `json:"status"`
	ErrorMessage string         `json:"error_message,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
}

type Chunk struct {
	ID         string `json:"id"`
	DocumentID string `json:"document_id"`
	OrgID      string `json:"org_id"`
	Content    string `json:"content"`
	Position   int    `json:"position"`
	TokenCount int    `json:"token_count"`
}

type Conversation struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Title     string    `json:"title,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type MessageRole string

const (
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
	MessageRoleSystem    MessageRole = "system"
)

type Message struct {
	ID             string      `json:"id"`
	ConversationID string      `json:"conversation_id"`
	Role           MessageRole `json:"role"`
	Content        string      `json:"content"`
	SourcesJSON    string      `json:"sources_json,omitempty"`
	CreatedAt      time.Time   `json:"created_at"`
}

type VectorResult struct {
	ChunkID string
	Score   float32
}
