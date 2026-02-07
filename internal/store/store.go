package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type Config struct {
	PermanentAllowKubectlSubcmds map[string][]string `json:"permanent_allow_kubectl_subcommands"` // context:subcommand pair
}

func LoadConfig() (Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return Config{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Config{}, nil
		}
		return Config{}, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func SaveConfig(cfg Config) error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func ConfigPath() (string, error) {
	root, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "konfirm", "config.json"), nil
}

func IsKubectlSubcommandAllowed(allowed map[string][]string, ctx string, subcommand string) bool {
	if subcommand == "" {
		return false
	}
	items := allowed[ctx]
	for _, item := range items {
		if item == subcommand {
			return true
		}
	}
	return false
}

func RemoveKubectlSubcommand(items []string, target string) []string {
	result := make([]string, 0, len(items))
	for _, item := range items {
		if item != target {
			result = append(result, item)
		}
	}
	return result
}
