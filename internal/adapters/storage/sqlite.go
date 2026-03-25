package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lex-rio/omni-knowledge-base-assistant/internal/domain"
	_ "modernc.org/sqlite"
)

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(dbPath string, migrationSQL []byte) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", dbPath+"?_pragma=journal_mode(wal)&_pragma=foreign_keys(on)")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	db.SetMaxOpenConns(1)

	if _, err := db.Exec(string(migrationSQL)); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrations: %w", err)
	}

	return &SQLiteStore{db: db}, nil
}

func (s *SQLiteStore) Close() error { return s.db.Close() }

// --- Organizations ---

func (s *SQLiteStore) CreateOrganization(ctx context.Context, org domain.Organization) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO organizations (id, name, created_at) VALUES (?, ?, ?)`,
		org.ID, org.Name, org.CreatedAt.Format(time.RFC3339))
	return err
}

func (s *SQLiteStore) GetOrganization(ctx context.Context, id string) (domain.Organization, error) {
	var org domain.Organization
	var createdAt string
	err := s.db.QueryRowContext(ctx,
		`SELECT id, name, created_at FROM organizations WHERE id = ?`, id,
	).Scan(&org.ID, &org.Name, &createdAt)
	if err != nil {
		return org, err
	}
	org.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	return org, nil
}

// --- Organization Profiles ---

func (s *SQLiteStore) UpsertProfile(ctx context.Context, p domain.OrganizationProfile) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO organization_profiles (org_id, contact_name, business_description, document_types, preferences, onboarding_completed, raw_answers, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, datetime('now'))
		ON CONFLICT(org_id) DO UPDATE SET
			contact_name = excluded.contact_name,
			business_description = excluded.business_description,
			document_types = excluded.document_types,
			preferences = excluded.preferences,
			onboarding_completed = excluded.onboarding_completed,
			raw_answers = excluded.raw_answers,
			updated_at = datetime('now')`,
		p.OrgID, p.ContactName, p.BusinessDescription, p.DocumentTypes, p.Preferences, p.OnboardingCompleted, p.RawAnswers)
	return err
}

func (s *SQLiteStore) GetProfile(ctx context.Context, orgID string) (domain.OrganizationProfile, error) {
	var p domain.OrganizationProfile
	var onboarded int
	err := s.db.QueryRowContext(ctx,
		`SELECT org_id, contact_name, business_description, document_types, preferences, onboarding_completed, raw_answers, updated_at
		FROM organization_profiles WHERE org_id = ?`, orgID,
	).Scan(&p.OrgID, &p.ContactName, &p.BusinessDescription, &p.DocumentTypes, &p.Preferences, &onboarded, &p.RawAnswers, &p.UpdatedAt)
	if err != nil {
		return p, err
	}
	p.OnboardingCompleted = onboarded == 1
	return p, nil
}

// --- Users ---

func (s *SQLiteStore) CreateUser(ctx context.Context, u domain.User) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO users (id, org_id, name, role, auth_token_hash, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		u.ID, u.OrgID, u.Name, u.Role, u.AuthTokenHash, u.CreatedAt.Format(time.RFC3339))
	return err
}

func (s *SQLiteStore) GetUser(ctx context.Context, id string) (domain.User, error) {
	var u domain.User
	var createdAt string
	err := s.db.QueryRowContext(ctx,
		`SELECT id, org_id, name, role, auth_token_hash, created_at FROM users WHERE id = ?`, id,
	).Scan(&u.ID, &u.OrgID, &u.Name, &u.Role, &u.AuthTokenHash, &createdAt)
	if err != nil {
		return u, err
	}
	u.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	return u, nil
}

func (s *SQLiteStore) GetUserByToken(ctx context.Context, tokenHash string) (domain.User, error) {
	var u domain.User
	var createdAt string
	err := s.db.QueryRowContext(ctx,
		`SELECT id, org_id, name, role, auth_token_hash, created_at FROM users WHERE auth_token_hash = ?`, tokenHash,
	).Scan(&u.ID, &u.OrgID, &u.Name, &u.Role, &u.AuthTokenHash, &createdAt)
	if err != nil {
		return u, err
	}
	u.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	return u, nil
}

// --- Documents ---

func (s *SQLiteStore) CreateDocument(ctx context.Context, doc domain.Document) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO documents (id, org_id, filename, mime_type, size_bytes, status, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		doc.ID, doc.OrgID, doc.Filename, doc.MimeType, doc.SizeBytes, doc.Status, doc.CreatedAt.Format(time.RFC3339))
	return err
}

func (s *SQLiteStore) UpdateDocumentStatus(ctx context.Context, id string, status domain.DocumentStatus, errMsg string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE documents SET status = ?, error_message = ? WHERE id = ?`,
		status, errMsg, id)
	return err
}

func (s *SQLiteStore) ListDocuments(ctx context.Context, orgID string) ([]domain.Document, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, org_id, filename, mime_type, size_bytes, status, error_message, created_at
		FROM documents WHERE org_id = ? ORDER BY created_at DESC`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []domain.Document
	for rows.Next() {
		var d domain.Document
		var createdAt string
		if err := rows.Scan(&d.ID, &d.OrgID, &d.Filename, &d.MimeType, &d.SizeBytes, &d.Status, &d.ErrorMessage, &createdAt); err != nil {
			return nil, err
		}
		d.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		docs = append(docs, d)
	}
	return docs, rows.Err()
}

func (s *SQLiteStore) DeleteDocument(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM documents WHERE id = ?`, id)
	return err
}

// --- Chunks ---

func (s *SQLiteStore) InsertChunks(ctx context.Context, chunks []domain.Chunk) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO chunks (id, document_id, org_id, content, position, token_count) VALUES (?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, c := range chunks {
		if _, err := stmt.ExecContext(ctx, c.ID, c.DocumentID, c.OrgID, c.Content, c.Position, c.TokenCount); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *SQLiteStore) GetChunksByIDs(ctx context.Context, ids []string) ([]domain.Chunk, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	query := `SELECT id, document_id, org_id, content, position, token_count FROM chunks WHERE id IN (`
	args := make([]any, len(ids))
	for i, id := range ids {
		if i > 0 {
			query += ","
		}
		query += "?"
		args[i] = id
	}
	query += ")"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chunks []domain.Chunk
	for rows.Next() {
		var c domain.Chunk
		if err := rows.Scan(&c.ID, &c.DocumentID, &c.OrgID, &c.Content, &c.Position, &c.TokenCount); err != nil {
			return nil, err
		}
		chunks = append(chunks, c)
	}
	return chunks, rows.Err()
}

func (s *SQLiteStore) DeleteChunksByDocument(ctx context.Context, documentID string) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id FROM chunks WHERE document_id = ?`, documentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	_, err = s.db.ExecContext(ctx, `DELETE FROM chunks WHERE document_id = ?`, documentID)
	return ids, err
}

// --- Conversations ---

func (s *SQLiteStore) CreateConversation(ctx context.Context, conv domain.Conversation) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO conversations (id, user_id, title, created_at) VALUES (?, ?, ?, ?)`,
		conv.ID, conv.UserID, conv.Title, conv.CreatedAt.Format(time.RFC3339))
	return err
}

func (s *SQLiteStore) ListConversations(ctx context.Context, userID string) ([]domain.Conversation, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, user_id, title, created_at FROM conversations WHERE user_id = ? ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var convs []domain.Conversation
	for rows.Next() {
		var c domain.Conversation
		var createdAt string
		if err := rows.Scan(&c.ID, &c.UserID, &c.Title, &createdAt); err != nil {
			return nil, err
		}
		c.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		convs = append(convs, c)
	}
	return convs, rows.Err()
}

// --- Messages ---

func (s *SQLiteStore) CreateMessage(ctx context.Context, msg domain.Message) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO messages (id, conversation_id, role, content, sources_json, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		msg.ID, msg.ConversationID, msg.Role, msg.Content, msg.SourcesJSON, msg.CreatedAt.Format(time.RFC3339))
	return err
}

func (s *SQLiteStore) ListMessages(ctx context.Context, conversationID string) ([]domain.Message, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, conversation_id, role, content, sources_json, created_at
		FROM messages WHERE conversation_id = ? ORDER BY created_at ASC`, conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []domain.Message
	for rows.Next() {
		var m domain.Message
		var createdAt string
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.Role, &m.Content, &m.SourcesJSON, &createdAt); err != nil {
			return nil, err
		}
		m.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		msgs = append(msgs, m)
	}
	return msgs, rows.Err()
}
