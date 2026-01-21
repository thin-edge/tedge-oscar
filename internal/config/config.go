package config

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"bytes"
	"strings"
	"text/template"

	"github.com/BurntSushi/toml"
	"oras.land/oras-go/v2/registry/remote/credentials"
)

//go:embed tedge-oscar.toml
var embeddedConfig []byte

type RegistryCredential struct {
	Registry string `toml:"registry" json:"registry" yaml:"registry"`
	Username string `toml:"username" json:"username" yaml:"username"`
	Password string `toml:"password" json:"password" yaml:"password"`
}

type Config struct {
	ImageDir            string               `toml:"image_dir" json:"image_dir" yaml:"image_dir"`
	DeployDir           string               `toml:"deploy_dir" json:"deploy_dir" yaml:"deploy_dir"`
	Registries          []RegistryCredential `toml:"registries" json:"registries" yaml:"registries"`
	UnexpandedImageDir  string               `toml:"-" json:"-" yaml:"-"`
	UnexpandedDeployDir string               `toml:"-" json:"-" yaml:"-"`
}

func DefaultConfigPath() string {
	if envPath := os.Getenv("TEDGE_OSCAR_CONFIG"); envPath != "" {
		return os.ExpandEnv(envPath)
	}

	if _, err := os.Stat("/etc/tedge/plugins/tedge-oscar.toml"); err == nil {
		return "/etc/tedge/plugins/tedge-oscar.toml"
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "./tedge-oscar.toml"
	}
	return filepath.Join(home, ".config", "tedge-oscar", "config.toml")
}

type TemplateContext struct {
	Env    map[string]string
	Mapper string
}

func NewTemplateContext(mapper string) *TemplateContext {
	env := make(map[string]string)
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if len(pair) == 2 {
			env[pair[0]] = pair[1]
		}
	}
	return &TemplateContext{
		Env:    env,
		Mapper: mapper,
	}
}

func (c *Config) evaluateTemplate(v string, mapper string) (string, error) {
	tmpl, err := template.New("eval").Parse(v)
	if err != nil {
		return v, err
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, NewTemplateContext(mapper))
	if err != nil {
		return v, err
	}
	return buf.String(), nil
}

func (c *Config) GetDeployDir(mapper string) (string, error) {
	return c.evaluateTemplate(c.DeployDir, mapper)
}

func expandEnvVars(s string) string {
	return os.ExpandEnv(s)
}

func (c *Config) Expand() {
	c.UnexpandedImageDir = c.ImageDir
	c.UnexpandedDeployDir = c.DeployDir
	c.ImageDir = expandEnvVars(c.ImageDir)
	c.DeployDir = expandEnvVars(c.DeployDir)
	for i := range c.Registries {
		c.Registries[i].Registry = expandEnvVars(c.Registries[i].Registry)
		c.Registries[i].Username = expandEnvVars(c.Registries[i].Username)
		c.Registries[i].Password = expandEnvVars(c.Registries[i].Password)
	}
}

func loadEmbeddedConfig() (*Config, error) {
	var cfg Config
	if err := toml.Unmarshal(embeddedConfig, &cfg); err != nil {
		return &cfg, fmt.Errorf("failed to load embedded config: %w", err)
	}
	cfg.Expand()
	return &cfg, nil
}

func LoadConfig(path string) (*Config, error) {
	// Set default tedge config dir
	if v := os.Getenv("TEDGE_CONFIG_DIR"); v == "" {
		os.Setenv("TEDGE_CONFIG_DIR", "/etc/tedge")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return loadEmbeddedConfig()
	}
	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, err
	}
	cfg.Expand()
	return &cfg, nil
}

func (c *Config) FindCredential(registry string) *RegistryCredential {
	// Prefer Docker credentials store if available
	username, password, err := LoadDockerCredentials(registry)
	if err == nil && username != "" && password != "" {
		return &RegistryCredential{
			Registry: registry,
			Username: username,
			Password: password,
		}
	}
	// Then try ORAS credentials store
	credStore, err := credentials.NewStore("", credentials.StoreOptions{})
	if err == nil {
		cred, err := credStore.Get(context.Background(), registry)
		if err == nil && cred.Username != "" && cred.Password != "" {
			return &RegistryCredential{
				Registry: registry,
				Username: cred.Username,
				Password: cred.Password,
			}
		}
	}
	// Fallback to config file
	for _, cred := range c.Registries {
		if cred.Registry == registry {
			return &cred
		}
	}
	return nil
}
