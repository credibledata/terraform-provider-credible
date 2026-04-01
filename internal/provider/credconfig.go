package provider

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// credConfigFile mirrors the ~/.cred JSON structure written by the Credible CLI.
type credConfigFile struct {
	ConfigPerEnvironment map[string]*credEnvironmentConfig `json:"configPerEnvironment"`
	ActiveEnvironment    string                            `json:"activeEnvironment"`
}

type credEnvironmentConfig struct {
	Organization          string `json:"organization,omitempty"`
	Project               string `json:"project,omitempty"`
	AccessToken           string `json:"accessToken,omitempty"`
	JwtAccessToken        string `json:"jwtAccessToken,omitempty"`
	IsServiceAccountToken *bool  `json:"isServiceAccountToken,omitempty"`
}

// credConfigAuth holds the resolved auth header and optional organization from the CLI config.
type credConfigAuth struct {
	AuthHeader   string
	Organization string
}

// readCredConfig reads ~/.cred (or the path in CREDIBLE_CONFIG_PATH) and returns
// the auth header and organization for the active environment.
func readCredConfig() (*credConfigAuth, error) {
	configPath := os.Getenv("CREDIBLE_CONFIG_PATH")
	if configPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("determining home directory: %w", err)
		}
		configPath = filepath.Join(home, ".cred")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", configPath, err)
	}

	var cfg credConfigFile
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", configPath, err)
	}

	env := cfg.ActiveEnvironment
	if env == "" {
		env = "production"
	}

	envCfg, ok := cfg.ConfigPerEnvironment[env]
	if !ok || envCfg == nil {
		return nil, fmt.Errorf("no configuration found for environment %q in %s", env, configPath)
	}

	token := envCfg.AccessToken
	if token == "" {
		return nil, fmt.Errorf("no access token found for environment %q in %s (run 'cred login' first)", env, configPath)
	}

	// Match CLI logic: isServiceAccountToken=true → "ApiKey", otherwise → "Bearer"
	authPrefix := "Bearer"
	if envCfg.IsServiceAccountToken != nil && *envCfg.IsServiceAccountToken {
		authPrefix = "ApiKey"
	}

	return &credConfigAuth{
		AuthHeader:   authPrefix + " " + token,
		Organization: envCfg.Organization,
	}, nil
}
