package models

import (
	"testing"
	"time"
)

func TestSourceValidation(t *testing.T) {
	tests := []struct {
		name    string
		source  Source
		wantErr bool
	}{
		{
			name: "valid source",
			source: Source{
				ID:          "fabric",
				Name:        "Fabric Patterns",
				URL:         "https://github.com/danielmiessler/fabric",
				LocalPath:   "/tmp/fabric",
				Format:      "fabric_pattern",
				PromptCount: 10,
			},
			wantErr: false,
		},
		{
			name: "invalid ID format",
			source: Source{
				ID:        "Fabric_Invalid",
				Name:      "Fabric Patterns",
				URL:       "https://github.com/danielmiessler/fabric",
				LocalPath: "/tmp/fabric",
				Format:    "fabric_pattern",
			},
			wantErr: true,
		},
		{
			name: "non-github URL",
			source: Source{
				ID:        "fabric",
				Name:      "Fabric Patterns",
				URL:       "https://gitlab.com/danielmiessler/fabric",
				LocalPath: "/tmp/fabric",
				Format:    "fabric_pattern",
			},
			wantErr: true,
		},
		{
			name: "invalid format",
			source: Source{
				ID:        "fabric",
				Name:      "Fabric Patterns",
				URL:       "https://github.com/danielmiessler/fabric",
				LocalPath: "/tmp/fabric",
				Format:    "invalid_format",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.source.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Source.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBookmarkValidation(t *testing.T) {
	tests := []struct {
		name     string
		bookmark Bookmark
		wantErr  bool
	}{
		{
			name: "valid bookmark",
			bookmark: Bookmark{
				Alias:      "review",
				PromptID:   "fabric:code-review",
				SourceID:   "fabric",
				PromptName: "code-review",
				Tags:       []string{"dev", "security"},
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
				UsageCount: 5,
			},
			wantErr: false,
		},
		{
			name: "invalid alias format (uppercase)",
			bookmark: Bookmark{
				Alias:      "Review",
				PromptID:   "fabric:code-review",
				SourceID:   "fabric",
				PromptName: "code-review",
			},
			wantErr: true,
		},
		{
			name: "reserved command as alias",
			bookmark: Bookmark{
				Alias:      "help",
				PromptID:   "fabric:code-review",
				SourceID:   "fabric",
				PromptName: "code-review",
			},
			wantErr: true,
		},
		{
			name: "invalid prompt ID format",
			bookmark: Bookmark{
				Alias:      "review",
				PromptID:   "invalid_prompt_id",
				SourceID:   "fabric",
				PromptName: "code-review",
			},
			wantErr: true,
		},
		{
			name: "mismatched prompt ID",
			bookmark: Bookmark{
				Alias:      "review",
				PromptID:   "fabric:wrong",
				SourceID:   "fabric",
				PromptName: "code-review",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.bookmark.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Bookmark.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name:    "valid default config",
			config:  *DefaultConfig(),
			wantErr: false,
		},
		{
			name: "invalid rate limit threshold (too low)",
			config: Config{
				Version: "1.0",
				GitHub: GitHubConfig{
					RateLimitWarningThreshold: 30,
				},
				Search: SearchConfig{
					MaxResults: 50,
				},
				Display: DisplayConfig{
					TableStyle: "rounded",
					DateFormat: "relative",
				},
				Cache: CacheConfig{},
			},
			wantErr: true,
		},
		{
			name: "invalid max results",
			config: Config{
				Version: "1.0",
				GitHub: GitHubConfig{
					RateLimitWarningThreshold: 80,
				},
				Search: SearchConfig{
					MaxResults: 5000,
				},
				Display: DisplayConfig{
					TableStyle: "rounded",
					DateFormat: "relative",
				},
				Cache: CacheConfig{},
			},
			wantErr: true,
		},
		{
			name: "invalid table style",
			config: Config{
				Version: "1.0",
				GitHub: GitHubConfig{
					RateLimitWarningThreshold: 80,
				},
				Search: SearchConfig{
					MaxResults: 50,
				},
				Display: DisplayConfig{
					TableStyle: "invalid",
					DateFormat: "relative",
				},
				Cache: CacheConfig{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
