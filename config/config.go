package config

import "errors"

type RepoConfig struct {
	Location string
}

func (config *RepoConfig) Set(section string, key string, value string) error {
	return nil
}

func (config *RepoConfig) Get(section string, key string) (string, error) {
	return "", errors.New("not implemented")
}
