// Package credentials provides secure credential storage for API keys.
// It uses the system keychain on macOS (Keychain), Linux (Secret Service),
// and falls back to encrypted file storage if no system keychain is available.
package credentials

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/zalando/go-keyring"
)

// Service name for the keychain
const ServiceName = "ai-manager"

// CredentialsStore manages API credentials securely
type CredentialsStore struct {
	mu      sync.RWMutex
	baseDir string
}

// Credential represents an API credential
type Credential struct {
	Model    string `json:"model"`
	Key      string `json:"key"`       // API key name (e.g., "ANTHROPIC_API_KEY")
	Value    string `json:"value"`     // The actual API key (never stored in plaintext in file)
	Source   string `json:"source"`    // "keychain", "env", "file"
	Provider string `json:"provider"`  // e.g., "anthropic", "minimax", "zhipu"
}

// CredentialInfo represents credential metadata (without the secret value)
type CredentialInfo struct {
	Model    string `json:"model"`
	Key      string `json:"key"`
	Source   string `json:"source"`
	Provider string `json:"provider"`
	Set      bool   `json:"set"` // Whether the credential is configured
}

// NewCredentialsStore creates a new credentials store
func NewCredentialsStore() (*CredentialsStore, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	baseDir := filepath.Join(home, ".ai-manager", "credentials")
	if err := os.MkdirAll(baseDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create credentials directory: %w", err)
	}

	return &CredentialsStore{
		baseDir: baseDir,
	}, nil
}

// Get retrieves a credential value by model and key name
func (s *CredentialsStore) Get(model, keyName string) (string, error) {
	// First, try to get from system keychain
	if s.isKeychainAvailable() {
		value, err := keyring.Get(ServiceName, s.keyringKey(model, keyName))
		if err == nil && value != "" {
			return value, nil
		}
	}

	// Fall back to encrypted file storage
	return s.getFromFile(model, keyName)
}

// Set stores a credential value securely
func (s *CredentialsStore) Set(model, keyName, value, provider string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Try system keychain first
	if s.isKeychainAvailable() {
		err := keyring.Set(ServiceName, s.keyringKey(model, keyName), value)
		if err == nil {
			return nil
		}
		// If keyring fails, continue to file storage
	}

	// Fall back to encrypted file storage
	return s.saveToFile(model, keyName, value, provider)
}

// Delete removes a credential
func (s *CredentialsStore) Delete(model, keyName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Delete from keychain
	if s.isKeychainAvailable() {
		//nolint:errcheck
		keyring.Delete(ServiceName, s.keyringKey(model, keyName))
	}

	// Delete from file
	return s.deleteFromFile(model, keyName)
}

// List returns all credentials (without values)
func (s *CredentialsStore) List() ([]CredentialInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []CredentialInfo

	// Note: go-keyring doesn't provide a List function, so we only list from files
	// The actual credential values are checked at runtime when needed

	// Check file storage for known credentials
	s.listFromFile(&results)

	return results, nil
}

// GetForModel returns all credentials for a specific model
func (s *CredentialsStore) GetForModel(model string) ([]CredentialInfo, error) {
	all, err := s.List()
	if err != nil {
		return nil, err
	}

	var results []CredentialInfo
	for _, c := range all {
		if c.Model == model {
			results = append(results, c)
		}
	}

	return results, nil
}

// HasCredentials checks if a model has any credentials set
func (s *CredentialsStore) HasCredentials(model string) (bool, error) {
	creds, err := s.GetForModel(model)
	if err != nil {
		return false, err
	}
	return len(creds) > 0, nil
}

// SetFromEnv marks an environment variable as the credential source
func (s *CredentialsStore) SetFromEnv(model, keyName, envVar, provider string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	filePath := filepath.Join(s.baseDir, model+"-"+keyName+".env.json")

	data := map[string]string{
		"source":   "env",
		"env_var":  envVar,
		"provider": provider,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, jsonData, 0600)
}

// GetEnvVar returns the environment variable name for a credential
func (s *CredentialsStore) GetEnvVar(model, keyName string) (string, bool, error) {
	filePath := filepath.Join(s.baseDir, model+"-"+keyName+".env.json")

	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", false, err
	}

	var info struct {
		Source  string `json:"source"`
		EnvVar  string `json:"env_var"`
	}
	if err := json.Unmarshal(data, &info); err != nil {
		return "", false, err
	}

	return info.EnvVar, info.Source == "env", nil
}

// ExportEnvExport generates export commands for credentials
func (s *CredentialsStore) ExportEnvExport(model string) (string, error) {
	creds, err := s.List()
	if err != nil {
		return "", err
	}

	var exportCmds []string
	for _, c := range creds {
		if c.Model == model && c.Source == "env" {
			envVar, _, _ := s.GetEnvVar(model, c.Key)
			if envVar != "" {
				exportCmds = append(exportCmds, fmt.Sprintf("export %s=$%s", c.Key, envVar))
			}
		}
	}

	return strings.Join(exportCmds, "\n"), nil
}

// keyringKey creates a unique key for the keychain
func (s *CredentialsStore) keyringKey(model, keyName string) string {
	return model + ":" + keyName
}

// isKeychainAvailable checks if system keychain is available
func (s *CredentialsStore) isKeychainAvailable() bool {
	// Check if we're on a supported platform
	switch runtime.GOOS {
	case "darwin", "linux":
		// Keyring should work on these platforms
		return true
	case "windows":
		// Windows keyring support is limited
		return false
	default:
		return false
	}
}

// file encryption key - in production, this should be derived from a user password
// For now, we use a platform-specific location to store the encryption key
func (s *CredentialsStore) getEncryptionKey() ([]byte, error) {
	keyFile := filepath.Join(s.baseDir, ".encryption_key")

	// Try to read existing key
	data, err := os.ReadFile(keyFile)
	if err == nil && len(data) == 32 {
		return data, nil
	}

	// Generate new key
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}

	// Save key with restricted permissions
	if err := os.WriteFile(keyFile, key, 0600); err != nil {
		return nil, err
	}

	return key, nil
}

// saveToFile saves a credential to an encrypted file
func (s *CredentialsStore) saveToFile(model, keyName, value, provider string) error {
	key, err := s.getEncryptionKey()
	if err != nil {
		return err
	}

	ciphertext, err := encrypt([]byte(value), key)
	if err != nil {
		return err
	}

	data := Credential{
		Model:    model,
		Key:      keyName,
		Value:    base64.StdEncoding.EncodeToString(ciphertext),
		Source:   "file",
		Provider: provider,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	filePath := filepath.Join(s.baseDir, model+"-"+keyName+".cred.json")
	return os.WriteFile(filePath, jsonData, 0600)
}

// getFromFile retrieves a credential from an encrypted file
func (s *CredentialsStore) getFromFile(model, keyName string) (string, error) {
	filePath := filepath.Join(s.baseDir, model+"-"+keyName+".cred.json")

	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	var cred Credential
	if err := json.Unmarshal(data, &cred); err != nil {
		return "", err
	}

	ciphertext, err := base64.StdEncoding.DecodeString(cred.Value)
	if err != nil {
		return "", err
	}

	key, err := s.getEncryptionKey()
	if err != nil {
		return "", err
	}

	plaintext, err := decrypt(ciphertext, key)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// deleteFromFile removes a credential file
func (s *CredentialsStore) deleteFromFile(model, keyName string) error {
	filePath := filepath.Join(s.baseDir, model+"-"+keyName+".cred.json")
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return err
	}

	envPath := filepath.Join(s.baseDir, model+"-"+keyName+".env.json")
	if err := os.Remove(envPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

// listFromFile lists credentials from file storage
func (s *CredentialsStore) listFromFile(results *[]CredentialInfo) {
	entries, _ := os.ReadDir(s.baseDir)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		if strings.HasSuffix(filename, ".cred.json") {
			parts := strings.TrimSuffix(filename, ".cred.json")
			modelKey := strings.SplitN(parts, "-", 2)
			if len(modelKey) == 2 {
				*results = append(*results, CredentialInfo{
					Model:  modelKey[0],
					Key:    modelKey[1],
					Source: "file",
					Set:    true,
				})
			}
		}
	}
}

// encrypt encrypts data using AES-256-GCM
func encrypt(plaintext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// decrypt decrypts data using AES-256-GCM
func decrypt(ciphertext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// DeleteAll removes all credentials
func (s *CredentialsStore) DeleteAll() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Clear files - keychain entries need to be cleared individually
	entries, _ := os.ReadDir(s.baseDir)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".cred.json") ||
			strings.HasSuffix(entry.Name(), ".env.json") ||
			entry.Name() == ".encryption_key" {
			os.Remove(filepath.Join(s.baseDir, entry.Name()))
		}
	}

	return nil
}
