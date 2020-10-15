package repo

type RepoConfig struct {
	location string
}

func (config *RepoConfig) Set(section string, key string, value string) error {
	return nil
}

func (config *RepoConfig) Get(section string, key string) (string, error) {
	return "", nil
}