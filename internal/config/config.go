// Package config is the pmx multi-project credential store. It mirrors the ddx
// config layer (atomic clicore writes, flag>env>file resolution) but carries
// the two Postmark token scopes per project. Neither token is mandatory at load
// time — a command requiring the server scope only needs server_token, and an
// account-scope command only needs account_token; the client raises a clear
// error at call time if the scope it needs is unset.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	cliconfig "github.com/nicolasacchi/clicore/config"
)

type Project struct {
	ServerToken  string `toml:"server_token,omitempty"`
	AccountToken string `toml:"account_token,omitempty"`
}

type Config struct {
	DefaultProject string              `toml:"default_project,omitempty"`
	Projects       map[string]*Project `toml:"projects,omitempty"`
}

func configFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "pmx", "config.toml"), nil
}

func loadConfigFile() (*Config, error) {
	path, err := configFilePath()
	if err != nil {
		return nil, err
	}
	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func saveConfigFile(cfg *Config) error {
	path, err := configFilePath()
	if err != nil {
		return err
	}
	// Atomic temp+rename (clicore) so an interrupted encode can't corrupt a
	// config that already holds credentials.
	return cliconfig.SaveTOML(path, cfg)
}

func resolveProject(cfg *Config, projectFlag string) *Project {
	if cfg == nil {
		return nil
	}
	if projectFlag != "" && cfg.Projects != nil {
		if p, ok := cfg.Projects[projectFlag]; ok {
			return p
		}
		return nil
	}
	if cfg.DefaultProject != "" && cfg.Projects != nil {
		if p, ok := cfg.Projects[cfg.DefaultProject]; ok {
			return p
		}
	}
	return nil
}

// Credentials holds the resolved Postmark tokens (either may be empty).
type Credentials struct {
	ServerToken  string
	AccountToken string
}

// LoadCredentials resolves both tokens from flag > env > config file. It never
// errors on a missing token — the per-scope check happens at request time.
func LoadCredentials(serverFlag, accountFlag, projectFlag string) (*Credentials, error) {
	var p *Project
	if cfg, err := loadConfigFile(); err == nil {
		p = resolveProject(cfg, projectFlag)
	}
	return &Credentials{
		ServerToken:  cliconfig.FirstNonEmpty(serverFlag, os.Getenv("POSTMARK_SERVER_TOKEN"), tokenOf(p, ScopeServerToken)),
		AccountToken: cliconfig.FirstNonEmpty(accountFlag, os.Getenv("POSTMARK_ACCOUNT_TOKEN"), tokenOf(p, ScopeAccountToken)),
	}, nil
}

// tokenKind distinguishes which project token to read.
type tokenKind int

const (
	ScopeServerToken tokenKind = iota
	ScopeAccountToken
)

func tokenOf(p *Project, kind tokenKind) string {
	if p == nil {
		return ""
	}
	if kind == ScopeServerToken {
		return p.ServerToken
	}
	return p.AccountToken
}

func AddProject(name, serverToken, accountToken string) error {
	cfg, err := loadConfigFile()
	if err != nil {
		cfg = &Config{}
	}
	if cfg.Projects == nil {
		cfg.Projects = make(map[string]*Project)
	}
	cfg.Projects[name] = &Project{ServerToken: serverToken, AccountToken: accountToken}
	if cfg.DefaultProject == "" {
		cfg.DefaultProject = name
	}
	return saveConfigFile(cfg)
}

func RemoveProject(name string) error {
	cfg, err := loadConfigFile()
	if err != nil {
		return fmt.Errorf("no config file found")
	}
	if cfg.Projects == nil {
		return fmt.Errorf("project %q not found", name)
	}
	if _, ok := cfg.Projects[name]; !ok {
		return fmt.Errorf("project %q not found", name)
	}
	delete(cfg.Projects, name)
	if cfg.DefaultProject == name {
		cfg.DefaultProject = ""
		for k := range cfg.Projects {
			cfg.DefaultProject = k
			break
		}
	}
	if len(cfg.Projects) == 0 {
		cfg.Projects = nil
	}
	return saveConfigFile(cfg)
}

func SetDefaultProject(name string) error {
	cfg, err := loadConfigFile()
	if err != nil {
		return fmt.Errorf("no config file found")
	}
	if cfg.Projects == nil {
		return fmt.Errorf("project %q not found", name)
	}
	if _, ok := cfg.Projects[name]; !ok {
		return fmt.Errorf("project %q not found", name)
	}
	cfg.DefaultProject = name
	return saveConfigFile(cfg)
}

func ListProjects() (*Config, error) {
	return loadConfigFile()
}

// MaskKey renders a non-revealing preview for `config list`.
func MaskKey(key string) string {
	if key == "" {
		return ""
	}
	if len(key) <= 10 {
		return "***"
	}
	return key[:8] + "***" + key[len(key)-4:]
}
