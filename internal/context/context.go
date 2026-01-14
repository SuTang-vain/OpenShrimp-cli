package context

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Conversation represents a single conversation
type Conversation struct {
	ID          string    `json:"id"`
	Tool        string    `json:"tool"`
	Model       string    `json:"model"`
	Title       string    `json:"title"`
	Summary     string    `json:"summary"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Messages    []Message `json:"messages"`
	Tags        []string  `json:"tags"`
	ProjectPath string    `json:"project_path"`
}

// Message represents a single message in a conversation
type Message struct {
	Role    string `json:"role"` // user, assistant, system
	Content string `json:"content"`
	Tokens  int    `json:"tokens,omitempty"`
}

// ProjectContext represents project-specific context
type ProjectContext struct {
	ID          string   `json:"id"`
	ProjectPath string   `json:"project_path"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	TechStack   []string `json:"tech_stack"`
	Keywords    []string `json:"keywords"`
	LastActive  time.Time `json:"last_active"`
	ContextFiles []string `json:"context_files"`
}

// ContextStats represents usage statistics
type ContextStats struct {
	TotalConversations int `json:"total_conversations"`
	TotalMessages      int `json:"total_messages"`
	StorageUsed        int64 `json:"storage_used_bytes"`
	ProjectsTracked    int `json:"projects_tracked"`
}

// Manager handles context storage and retrieval
type Manager struct {
	dbPath string
	db     *sql.DB
}

// NewManager creates a new context manager
func NewManager() (*Manager, error) {
	home, _ := os.UserHomeDir()
	dataDir := filepath.Join(home, ".ai-manager", "data")
	dbPath := filepath.Join(dataDir, "context.db")

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	m := &Manager{
		dbPath: dbPath,
		db:     db,
	}

	if err := m.initDB(); err != nil {
		return nil, err
	}

	return m, nil
}

func (m *Manager) initDB() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS conversations (
			id TEXT PRIMARY KEY,
			tool TEXT NOT NULL,
			model TEXT NOT NULL,
			title TEXT,
			summary TEXT,
			messages TEXT NOT NULL,
			tags TEXT,
			project_path TEXT,
			created_at DATETIME,
			updated_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS projects (
			id TEXT PRIMARY KEY,
			project_path TEXT UNIQUE NOT NULL,
			name TEXT NOT NULL,
			description TEXT,
			tech_stack TEXT,
			keywords TEXT,
			context_files TEXT,
			last_active DATETIME
		)`,
		`CREATE INDEX IF NOT EXISTS idx_conversations_tool ON conversations(tool)`,
		`CREATE INDEX IF NOT EXISTS idx_conversations_project ON conversations(project_path)`,
		`CREATE INDEX IF NOT EXISTS idx_projects_path ON projects(project_path)`,
	}

	for _, q := range queries {
		if _, err := m.db.Exec(q); err != nil {
			return err
		}
	}

	return nil
}

// AddConversation saves a new conversation
func (m *Manager) AddConversation(conv *Conversation) error {
	if conv.ID == "" {
		conv.ID = generateID()
	}
	if conv.CreatedAt.IsZero() {
		conv.CreatedAt = time.Now()
	}
	conv.UpdatedAt = time.Now()

	messagesJSON, _ := json.Marshal(conv.Messages)
	tagsJSON, _ := json.Marshal(conv.Tags)

	query := `INSERT OR REPLACE INTO conversations
		(id, tool, model, title, summary, messages, tags, project_path, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := m.db.Exec(query,
		conv.ID, conv.Tool, conv.Model, conv.Title, conv.Summary,
		string(messagesJSON), string(tagsJSON), conv.ProjectPath,
		conv.CreatedAt, conv.UpdatedAt)

	return err
}

// GetConversation retrieves a conversation by ID
func (m *Manager) GetConversation(id string) (*Conversation, error) {
	query := `SELECT id, tool, model, title, summary, messages, tags, project_path, created_at, updated_at
		FROM conversations WHERE id = ?`

	var conv Conversation
	var messagesJSON, tagsJSON string

	err := m.db.QueryRow(query, id).Scan(
		&conv.ID, &conv.Tool, &conv.Model, &conv.Title, &conv.Summary,
		&messagesJSON, &tagsJSON, &conv.ProjectPath,
		&conv.CreatedAt, &conv.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	//nolint:errcheck
	json.Unmarshal([]byte(messagesJSON), &conv.Messages)
	//nolint:errcheck
	json.Unmarshal([]byte(tagsJSON), &conv.Tags)

	return &conv, nil
}

// ListConversations returns conversations with optional filters
func (m *Manager) ListConversations(tool, projectPath string, limit int) ([]*Conversation, error) {
	query := `SELECT id, tool, model, title, summary, messages, tags, project_path, created_at, updated_at
		FROM conversations WHERE 1=1`

	args := []interface{}{}
	if tool != "" {
		query += " AND tool = ?"
		args = append(args, tool)
	}
	if projectPath != "" {
		query += " AND project_path = ?"
		args = append(args, projectPath)
	}

	query += " ORDER BY updated_at DESC"
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := m.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conversations []*Conversation
	for rows.Next() {
		var conv Conversation
		var messagesJSON, tagsJSON string

		//nolint:errcheck
		rows.Scan(&conv.ID, &conv.Tool, &conv.Model, &conv.Title, &conv.Summary,
			&messagesJSON, &tagsJSON, &conv.ProjectPath,
			&conv.CreatedAt, &conv.UpdatedAt)

		//nolint:errcheck
		json.Unmarshal([]byte(messagesJSON), &conv.Messages)
		//nolint:errcheck
		json.Unmarshal([]byte(tagsJSON), &conv.Tags)

		conversations = append(conversations, &conv)
	}

	return conversations, nil
}

// DeleteConversation removes a conversation
func (m *Manager) DeleteConversation(id string) error {
	_, err := m.db.Exec("DELETE FROM conversations WHERE id = ?", id)
	return err
}

// AddProject adds or updates a project context
func (m *Manager) AddProject(proj *ProjectContext) error {
	if proj.ID == "" {
		proj.ID = generateID()
	}
	if proj.LastActive.IsZero() {
		proj.LastActive = time.Now()
	}

	techStackJSON, _ := json.Marshal(proj.TechStack)
	keywordsJSON, _ := json.Marshal(proj.Keywords)
	contextFilesJSON, _ := json.Marshal(proj.ContextFiles)

	query := `INSERT OR REPLACE INTO projects
		(id, project_path, name, description, tech_stack, keywords, context_files, last_active)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := m.db.Exec(query,
		proj.ID, proj.ProjectPath, proj.Name, proj.Description,
		string(techStackJSON), string(keywordsJSON), string(contextFilesJSON),
		proj.LastActive)

	return err
}

// GetProject retrieves a project by path
func (m *Manager) GetProject(projectPath string) (*ProjectContext, error) {
	query := `SELECT id, project_path, name, description, tech_stack, keywords, context_files, last_active
		FROM projects WHERE project_path = ?`

	var proj ProjectContext
	var techStackJSON, keywordsJSON, contextFilesJSON string

	err := m.db.QueryRow(query, projectPath).Scan(
		&proj.ID, &proj.ProjectPath, &proj.Name, &proj.Description,
		&techStackJSON, &keywordsJSON, &contextFilesJSON, &proj.LastActive)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(techStackJSON), &proj.TechStack)
	json.Unmarshal([]byte(keywordsJSON), &proj.Keywords)
	json.Unmarshal([]byte(contextFilesJSON), &proj.ContextFiles)

	return &proj, nil
}

// ListProjects returns all tracked projects
func (m *Manager) ListProjects() ([]*ProjectContext, error) {
	rows, err := m.db.Query(`SELECT id, project_path, name, description, tech_stack, keywords, context_files, last_active
		FROM projects ORDER BY last_active DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*ProjectContext
	for rows.Next() {
		var proj ProjectContext
		var techStackJSON, keywordsJSON, contextFilesJSON string

		//nolint:errcheck
		rows.Scan(&proj.ID, &proj.ProjectPath, &proj.Name, &proj.Description,
			&techStackJSON, &keywordsJSON, &contextFilesJSON, &proj.LastActive)

		//nolint:errcheck
		json.Unmarshal([]byte(techStackJSON), &proj.TechStack)
		//nolint:errcheck
		json.Unmarshal([]byte(keywordsJSON), &proj.Keywords)
		//nolint:errcheck
		json.Unmarshal([]byte(contextFilesJSON), &proj.ContextFiles)

		projects = append(projects, &proj)
	}

	return projects, nil
}

// GetStats returns context storage statistics
func (m *Manager) GetStats() (*ContextStats, error) {
	var stats ContextStats

	// Count conversations
	if err := m.db.QueryRow("SELECT COUNT(*) FROM conversations").Scan(&stats.TotalConversations); err != nil {
		return nil, err
	}

	// Count messages (approximate)
	var messageCounts int
	rows, err := m.db.Query("SELECT messages FROM conversations")
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var messagesJSON string
		//nolint:errcheck
		rows.Scan(&messagesJSON)
		var msgs []Message
		//nolint:errcheck
		json.Unmarshal([]byte(messagesJSON), &msgs)
		messageCounts += len(msgs)
	}
	rows.Close()
	stats.TotalMessages = messageCounts

	// Count projects
	if err := m.db.QueryRow("SELECT COUNT(*) FROM projects").Scan(&stats.ProjectsTracked); err != nil {
		return nil, err
	}

	// Get storage size
	info, err := os.Stat(m.dbPath)
	if err == nil {
		stats.StorageUsed = info.Size()
	}

	return &stats, nil
}

// Close closes the database connection
func (m *Manager) Close() error {
	return m.db.Close()
}

// SearchConversations searches conversations by content
func (m *Manager) SearchConversations(query string) ([]*Conversation, error) {
	searchQuery := `%` + query + `%`
	rows, err := m.db.Query(`SELECT id, tool, model, title, summary, messages, tags, project_path, created_at, updated_at
		FROM conversations WHERE title LIKE ? OR summary LIKE ? OR messages LIKE ?
		ORDER BY updated_at DESC LIMIT 50`, searchQuery, searchQuery, searchQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conversations []*Conversation
	for rows.Next() {
		var conv Conversation
		var messagesJSON, tagsJSON string

		rows.Scan(&conv.ID, &conv.Tool, &conv.Model, &conv.Title, &conv.Summary,
			&messagesJSON, &tagsJSON, &conv.ProjectPath,
			&conv.CreatedAt, &conv.UpdatedAt)

		json.Unmarshal([]byte(messagesJSON), &conv.Messages)
		json.Unmarshal([]byte(tagsJSON), &conv.Tags)

		conversations = append(conversations, &conv)
	}

	return conversations, nil
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
