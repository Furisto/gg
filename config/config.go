package config

import (
	"errors"
	"gopkg.in/ini.v1"
	"os"
	"path/filepath"
	"runtime"
)

var (
	ErrUnavailable  = errors.New("configuration unavailable")
	ErrUnknownKey   = errors.New("unknown key")
	ErrCannotSetKey = errors.New("cannot set key")
)

func LocateSystemConfig() string {
	var systemPath string

	switch runtime.GOOS {
	case "windows":
		profile := os.Getenv("ALLUSERSPROFILE")
		systemPath = filepath.Join(profile, "Git", "config")
	case "linux":
		systemPath = "/etc/gitconfig"
	default:
		panic("operating system target not supported")
	}

	if _, err := os.Stat(systemPath); os.IsNotExist(err) {
		return ""
	}
	return systemPath
}

func LocateGlobalConfig() string {
	var globalPath string

	switch runtime.GOOS {
	case "windows":
		fallthrough
	case "linux":
		profile, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		globalPath = filepath.Join(profile, ".gitconfig")
	default:
		panic("operating system target not supported")
	}

	if _, err := os.Stat(globalPath); os.IsNotExist(err) {
		return ""
	}
	return globalPath
}

type Config interface {
	Set(section string, key string, value string) error
	Get(section string, key string) (string, error)
}

type IniConfig struct {
	document *ini.File
	location string
}

func NewIniConfig(path string) (*IniConfig, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		cfg := ini.Empty()
		if err := cfg.SaveTo(path); err != nil {
			return nil, err
		}
	}

	return &IniConfig{
		location: path,
	}, nil
}

func (cfg *IniConfig) Set(section string, key string, value string) error {
	if err := cfg.ensureDocument(); err != nil {
		return ErrUnavailable
	}
	s, err := cfg.document.NewSection(section)
	if err != nil {
		return ErrCannotSetKey
	}

	_, err = s.NewKey(key, value)
	if err != nil {
		return ErrCannotSetKey
	}

	if err := cfg.document.SaveTo(cfg.location); err != nil {
		return ErrCannotSetKey
	}

	return nil
}

func (cfg *IniConfig) Get(section string, key string) (string, error) {
	if err := cfg.ensureDocument(); err != nil {
		return "", ErrUnavailable
	}
	s, err := cfg.document.GetSection(section)
	if err != nil {
		return "", ErrUnknownKey
	}

	k, err := s.GetKey(key)
	if err != nil {
		return "", ErrUnknownKey
	}
	return k.Value(), err
}

func (cfg *IniConfig) ensureDocument() error {
	if cfg.document != nil {
		return nil
	}

	doc, err := ini.Load(cfg.location)
	if err != nil {
		return err
	}

	cfg.document = doc
	return nil
}

type UnionConfig struct {
	configs []Config
}

func (cfg *UnionConfig) Set(section string, key string, value string) error {
	for _, c := range cfg.configs {
		if err := c.Set(section, key, value); err == nil {
			return nil
		}
	}

	return ErrCannotSetKey
}

func (cfg *UnionConfig) Get(section string, key string) (string, error) {
	for _, c := range cfg.configs {
		if value, err := c.Get(section, key); err == nil {
			return value, nil
		}
	}

	return "", ErrUnknownKey
}

type NilConfig struct{}

func (cfg *NilConfig) Set(section, key, value string) error {
	return nil
}

func (cfg *NilConfig) Get(section, key string) (string, error) {
	return "", ErrUnknownKey
}

type InMemoryConfig struct {
	config map[string]map[string]string
}

func NewInMemoryConfig() InMemoryConfig {
	return InMemoryConfig{
		config: make(map[string]map[string]string),
	}
}

func (cfg *InMemoryConfig) Set(section, key, value string) error {
	s, exists := cfg.config[section]
	if !exists {
		cfg.config[section] = make(map[string]string)
		s = cfg.config[section]
	}

	s[key] = value
	return nil
}

func (cfg *InMemoryConfig) Get(section, key string) (string, error) {
	s, exists := cfg.config[section]
	if !exists {
		return "", ErrUnknownKey
	}

	v, exists := s[key]
	if !exists {
		return "", ErrUnknownKey
	}

	return v, nil
}

type ConfigBuilder struct {
	configs []Config
}

func CreateDefaultConfigBuilder(repoPath string) (*ConfigBuilder, error) {
	cb := ConfigBuilder{}
	cfgPaths := []string{repoPath, LocateGlobalConfig(), LocateSystemConfig()}

	for _, path := range cfgPaths {
		if path == "" {
			return nil, errors.New("empty path")
		}

		if err := cb.AddIniFile(path); err != nil {
			return nil, err
		}
	}

	return &cb, nil
}

func (cb *ConfigBuilder) AddIniFile(path string) error {
	cfg, err := NewIniConfig(path)
	if err != nil {
		return err
	}

	cb.configs = append(cb.configs, cfg)
	return nil
}

func (cb *ConfigBuilder) AddInMemory(initial map[string]map[string]string) error {
	cfg := InMemoryConfig{
		config: initial,
	}

	cb.configs = append(cb.configs, &cfg)
	return nil
}

func (cb *ConfigBuilder) Build() Config {
	return &UnionConfig{
		configs: cb.configs,
	}
}
