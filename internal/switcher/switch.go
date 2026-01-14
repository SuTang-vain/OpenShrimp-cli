package switcher

import (
	"fmt"
	"os"

	"ai-manager/internal/config"
	"ai-manager/internal/credentials"
	"ai-manager/internal/models"
	"ai-manager/internal/utils"
)

type Switcher struct {
	cfg *config.Config
}

func NewSwitcher(cfg *config.Config) *Switcher {
	return &Switcher{cfg: cfg}
}

func (s *Switcher) ListModels() []string {
	models := make([]string, 0, len(s.cfg.Models))
	for name := range s.cfg.Models {
		models = append(models, name)
	}
	return models
}

func (s *Switcher) GetCurrentModel() string {
	return s.cfg.Defaults.Model
}

func (s *Switcher) GetModel(name string) (config.Model, bool) {
	model, ok := s.cfg.Models[name]
	return model, ok
}

func (s *Switcher) SwitchTo(modelName string) error {
	if _, ok := s.cfg.Models[modelName]; !ok {
		return fmt.Errorf("model not found: %s", modelName)
	}

	s.cfg.Defaults.Model = modelName
	return nil
}

func (s *Switcher) ShowModelInfo(modelName string) error {
	model, ok := s.cfg.Models[modelName]
	if !ok {
		return fmt.Errorf("model not found: %s", modelName)
	}

	fmt.Printf("=== Model: %s ===\n", model.Name)
	fmt.Printf("Provider: %s\n", model.Provider)
	fmt.Printf("API Endpoint: %s\n", model.APIEndpoint)
	fmt.Printf("Model ID: %s\n", model.ModelID)

	if len(model.Environment) > 0 {
		fmt.Println("\nEnvironment Variables:")
		for key, value := range model.Environment {
			if isSecret(key) {
				fmt.Printf("  %s: [set]\n", key)
			} else {
				fmt.Printf("  %s: %s\n", key, value)
			}
		}
	}

	return nil
}

func isSecret(key string) bool {
	secrets := []string{"API_KEY", "APIKEY", "KEY", "TOKEN", "SECRET"}
	for _, s := range secrets {
		if key == s {
			return true
		}
	}
	return false
}

func (s *Switcher) ShowCurrentConfig() error {
	currentModel := s.cfg.Defaults.Model
	fmt.Printf("Current model: %s\n\n", currentModel)

	model, ok := s.cfg.Models[currentModel]
	if !ok {
		return fmt.Errorf("current model configuration not found")
	}

	fmt.Println("=== Current Configuration ===")
	fmt.Printf("Provider: %s\n", model.Provider)
	fmt.Printf("Model ID: %s\n", model.ModelID)

	fmt.Println("\n=== All Available Models ===")
	for name, m := range s.cfg.Models {
		prefix := "  "
		if name == currentModel {
			prefix = "* "
		}
		fmt.Printf("%s%s (%s)\n", prefix, name, m.Name)
	}

	return nil
}

func (s *Switcher) ExportEnvConfig(modelName string) (string, error) {
	model, ok := s.cfg.Models[modelName]
	if !ok {
		return "", fmt.Errorf("model not found: %s", modelName)
	}

	envOutput := fmt.Sprintf("# AI Model Configuration: %s\n", model.Name)
	envOutput += fmt.Sprintf("export AI_PROVIDER=%s\n", model.Provider)
	envOutput += fmt.Sprintf("export AI_MODEL=%s\n", model.ModelID)
	envOutput += fmt.Sprintf("export AI_API_ENDPOINT=%s\n", model.APIEndpoint)

	for key, value := range model.Environment {
		envOutput += fmt.Sprintf("export %s=%s\n", key, value)
	}

	return envOutput, nil
}

func (s *Switcher) Save() error {
	return config.Save(s.cfg, config.GetDefaultConfigPath())
}

type SwitchResult struct {
	Model       string `json:"model"`
	Provider    string `json:"provider"`
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	EnvExported string `json:"env_exported,omitempty"`
}

func (s *Switcher) PerformSwitch(modelName string, exportEnv bool) (*SwitchResult, error) {
	result := &SwitchResult{
		Model: modelName,
	}

	model, ok := s.cfg.Models[modelName]
	if !ok {
		result.Success = false
		result.Message = fmt.Sprintf("Model not found: %s", modelName)
		return result, fmt.Errorf("model not found: %s", modelName)
	}

	result.Provider = model.Provider

	if err := s.SwitchTo(modelName); err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("Failed to switch: %v", err)
		return result, err
	}

	if err := s.Save(); err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("Failed to save config: %v", err)
		return result, err
	}

	result.Success = true
	result.Message = fmt.Sprintf("Switched to %s (%s)", model.Name, model.Provider)

	if exportEnv {
		envConfig, _ := s.ExportEnvConfig(modelName)
		result.EnvExported = envConfig
	}

	return result, nil
}

func ValidateAPIKey(modelName string) (bool, error) {
	envFile := utils.ExpandPath("~/.ai-manager/.env." + modelName)

	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		return false, nil
	}

	return true, nil
}

func FormatBytes(bytes int64) string {
	return models.FormatBytes(bytes)
}

// GetCredentialStore returns a new credential store
func GetCredentialStore() (*credentials.CredentialsStore, error) {
	return credentials.NewCredentialsStore()
}

// SetModelCredential sets a credential for a model
func SetModelCredential(model, keyName, value, provider string) error {
	store, err := credentials.NewCredentialsStore()
	if err != nil {
		return err
	}
	return store.Set(model, keyName, value, provider)
}

// GetModelCredential retrieves a credential for a model
func GetModelCredential(model, keyName string) (string, error) {
	store, err := credentials.NewCredentialsStore()
	if err != nil {
		return "", err
	}
	return store.Get(model, keyName)
}

// DeleteModelCredential deletes a credential for a model
func DeleteModelCredential(model, keyName string) error {
	store, err := credentials.NewCredentialsStore()
	if err != nil {
		return err
	}
	return store.Delete(model, keyName)
}

// ListModelCredentials lists all credentials for a model
func ListModelCredentials(model string) ([]credentials.CredentialInfo, error) {
	store, err := credentials.NewCredentialsStore()
	if err != nil {
		return nil, err
	}
	return store.GetForModel(model)
}

// CheckModelCredentials checks if a model has required credentials
func CheckModelCredentials(model string) (map[string]bool, error) {
	store, err := credentials.NewCredentialsStore()
	if err != nil {
		return nil, err
	}

	creds, err := store.GetForModel(model)
	if err != nil {
		return nil, err
	}

	result := make(map[string]bool)
	for _, c := range creds {
		result[c.Key] = c.Set
	}
	return result, nil
}

// SetEnvCredential sets an environment variable reference for a credential
func SetEnvCredential(model, keyName, envVar, provider string) error {
	store, err := credentials.NewCredentialsStore()
	if err != nil {
		return err
	}
	return store.SetFromEnv(model, keyName, envVar, provider)
}

// GetEnvVarForCredential gets the environment variable for a credential
func GetEnvVarForCredential(model, keyName string) (string, bool, error) {
	store, err := credentials.NewCredentialsStore()
	if err != nil {
		return "", false, err
	}
	return store.GetEnvVar(model, keyName)
}

// HasModelCredentials checks if a model has any credentials set
func HasModelCredentials(model string) (bool, error) {
	store, err := credentials.NewCredentialsStore()
	if err != nil {
		return false, err
	}
	return store.HasCredentials(model)
}
