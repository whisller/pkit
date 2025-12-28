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
	now := time.Now()
	tests := []struct {
		name     string
		bookmark Bookmark
		wantErr  bool
	}{
		{
			name: "valid bookmark",
			bookmark: Bookmark{
				PromptID:   "fabric:code-review",
				Notes:      "My code review notes",
				CreatedAt:  now,
				UpdatedAt:  now,
				UsageCount: 5,
			},
			wantErr: false,
		},
		{
			name: "valid bookmark without notes",
			bookmark: Bookmark{
				PromptID:   "fabric:summarize",
				CreatedAt:  now,
				UpdatedAt:  now,
				UsageCount: 0,
			},
			wantErr: false,
		},
		{
			name: "invalid bookmark with empty prompt_id",
			bookmark: Bookmark{
				PromptID:   "",
				Notes:      "Some notes",
				CreatedAt:  now,
				UpdatedAt:  now,
				UsageCount: 0,
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
