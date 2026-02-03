package config

import (
	"encoding/json"
	"os"
	"path/filepath"
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

func Load() (*Config, error) {
	token := os.Getenv("GITHUB_TOKEN")

	reposConfig := loadReposConfig()

	return &Config{
		GitHubToken: token,
		ReposConfig: reposConfig,
	}, nil
}

func loadReposConfig() ReposConfig {
	defaultConfig := ReposConfig{
		Default: BranchConfig{
			ProdBranch: "main",
			DevBranch:  "develop",
		},
		Repositories: make(map[string]BranchConfig),
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return defaultConfig
	}

	configPath := filepath.Join(homeDir, ".mcp-repo-monitor", "repos.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return defaultConfig
	}

	var config ReposConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return defaultConfig
	}

	if config.Repositories == nil {
		config.Repositories = make(map[string]BranchConfig)
	}

	return config
}

func (c *Config) GetBranchConfig(repoFullName string) BranchConfig {
	if bc, ok := c.ReposConfig.Repositories[repoFullName]; ok {
		return bc
	}
	return c.ReposConfig.Default
}
