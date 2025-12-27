package models

// Config represents user configuration.
// Full implementation will be added in Phase 2 (Foundation).
type Config struct {
	Version    string   `yaml:"version" json:"version"`
	Sources    []Source `yaml:"sources" json:"sources"`
	GitHubAuth struct {
		UseAuth bool `yaml:"use_auth" json:"use_auth"`
	} `yaml:"github_auth" json:"github_auth"`
	IndexPath string `yaml:"index_path" json:"index_path"`
}
