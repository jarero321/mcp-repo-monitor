package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/carlos/mcp-repo-monitor/internal/infrastructure/logging"
	"github.com/joho/godotenv"
)

// Typed errors for config validation.
var (
	ErrInvalidJSON       = errors.New("invalid JSON in repos.json")
	ErrMissingDefault    = errors.New("missing default branch configuration")
	ErrInvalidBranchName = errors.New("invalid branch name")
)

type BranchConfig struct {
	ProdBranch string `json:"prod_branch"`
	DevBranch  string `json:"dev_branch"`
}

type ReposConfig struct {
	Default      BranchConfig            `json:"default"`
	Repositories map[string]BranchConfig `json:"repositories"`
}

type Config struct {
	GitHubToken string
	ReposConfig ReposConfig
}

func Load(logger *logging.Logger) (*Config, error) {
	loadEnvFiles(logger)

	token := os.Getenv("GITHUB_TOKEN")
	reposConfig, err := loadReposConfig(logger)
	if err != nil {
		return nil, err
	}

	logger.Info("config loaded successfully",
		"has_token", token != "",
		"repos_configured", len(reposConfig.Repositories),
	)

	return &Config{
		GitHubToken: token,
		ReposConfig: reposConfig,
	}, nil
}

func loadEnvFiles(logger *logging.Logger) {
	if err := godotenv.Load(); err == nil {
		logger.Debug("loaded .env from current directory")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}

	configDir := filepath.Join(homeDir, ".mcp-repo-monitor")
	envPath := filepath.Join(configDir, ".env")
	if err := godotenv.Load(envPath); err == nil {
		logger.Debug("loaded .env from config directory", "path", envPath)
	}
}

func loadReposConfig(logger *logging.Logger) (ReposConfig, error) {
	defaultConfig := ReposConfig{
		Default: BranchConfig{
			ProdBranch: "main",
			DevBranch:  "develop",
		},
		Repositories: make(map[string]BranchConfig),
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Debug("could not get home directory, using defaults")
		return defaultConfig, nil
	}

	configPath := filepath.Join(homeDir, ".mcp-repo-monitor", "repos.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Debug("repos.json not found, using defaults", "path", configPath)
			return defaultConfig, nil
		}
		return defaultConfig, fmt.Errorf("failed to read repos.json: %w", err)
	}

	var config ReposConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return ReposConfig{}, fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}

	// Validate default branch config
	if err := validateBranchConfig("default", config.Default); err != nil {
		return ReposConfig{}, err
	}

	// Validate repository-specific configs
	for repoName, branchConfig := range config.Repositories {
		if err := validateBranchConfig(repoName, branchConfig); err != nil {
			return ReposConfig{}, err
		}
	}

	if config.Repositories == nil {
		config.Repositories = make(map[string]BranchConfig)
	}

	logger.Info("repos.json loaded",
		"path", configPath,
		"default_prod", config.Default.ProdBranch,
		"default_dev", config.Default.DevBranch,
		"custom_repos", len(config.Repositories),
	)

	return config, nil
}

func validateBranchConfig(name string, bc BranchConfig) error {
	if bc.ProdBranch == "" {
		return fmt.Errorf("%w: prod_branch is empty for %s", ErrInvalidBranchName, name)
	}
	if bc.DevBranch == "" {
		return fmt.Errorf("%w: dev_branch is empty for %s", ErrInvalidBranchName, name)
	}

	// Basic branch name validation (no spaces, special chars)
	if strings.ContainsAny(bc.ProdBranch, " \t\n") {
		return fmt.Errorf("%w: prod_branch contains whitespace for %s", ErrInvalidBranchName, name)
	}
	if strings.ContainsAny(bc.DevBranch, " \t\n") {
		return fmt.Errorf("%w: dev_branch contains whitespace for %s", ErrInvalidBranchName, name)
	}

	return nil
}

func (c *Config) GetBranchConfig(repoFullName string) BranchConfig {
	if bc, ok := c.ReposConfig.Repositories[repoFullName]; ok {
		return bc
	}
	return c.ReposConfig.Default
}
