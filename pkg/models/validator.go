package models

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var (
	// validate is the shared validator instance
	validate *validator.Validate

	// Custom regex patterns
	sourceIDRegex   = regexp.MustCompile(`^[a-z0-9-]+(/[a-z0-9-]+)?$`)
	promptNameRegex = regexp.MustCompile(`^[a-z0-9_-]+$`)
	aliasRegex      = regexp.MustCompile(`^[a-z0-9_-]+$`)
	promptIDRegex   = regexp.MustCompile(`^[a-z0-9-]+(/[a-z0-9-]+)?:[a-z0-9_-]+$`)
	shaRegex        = regexp.MustCompile(`^[a-f0-9]{40}$`)
	tagRegex        = regexp.MustCompile(`^[a-z0-9_-]+$`)
)

func init() {
	validate = validator.New()

	// Register custom validators (errors would cause panic in init, acceptable)
	_ = validate.RegisterValidation("source_id", validateSourceID)
	_ = validate.RegisterValidation("github_url", validateGitHubURL)
	_ = validate.RegisterValidation("git_sha", validateGitSHA)
	_ = validate.RegisterValidation("prompt_name", validatePromptName)
	_ = validate.RegisterValidation("prompt_id", validatePromptID)
	_ = validate.RegisterValidation("alias", validateAlias)
	_ = validate.RegisterValidation("tag", validateTag)
	_ = validate.RegisterValidation("table_style", validateTableStyle)
	_ = validate.RegisterValidation("date_format", validateDateFormat)
}

// validateSourceID validates source ID format: lowercase alphanumeric with hyphens
func validateSourceID(fl validator.FieldLevel) bool {
	return sourceIDRegex.MatchString(fl.Field().String())
}

// validateGitHubURL validates that the URL is a GitHub repository URL
func validateGitHubURL(fl validator.FieldLevel) bool {
	urlStr := fl.Field().String()
	if urlStr == "" {
		return true // Let 'required' handle empty check
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	return u.Host == "github.com"
}

// validateGitSHA validates git commit SHA: 40 hex characters
func validateGitSHA(fl validator.FieldLevel) bool {
	sha := fl.Field().String()
	if sha == "" {
		return true // Optional field
	}
	return shaRegex.MatchString(sha)
}

// validatePromptName validates prompt name: lowercase alphanumeric with hyphens/underscores
func validatePromptName(fl validator.FieldLevel) bool {
	return promptNameRegex.MatchString(fl.Field().String())
}

// validatePromptID validates prompt ID format: <source_id>:<prompt_name>
func validatePromptID(fl validator.FieldLevel) bool {
	return promptIDRegex.MatchString(fl.Field().String())
}

// validateAlias validates alias format: lowercase alphanumeric with hyphens/underscores
func validateAlias(fl validator.FieldLevel) bool {
	return aliasRegex.MatchString(fl.Field().String())
}

// validateTag validates tag format: lowercase alphanumeric with hyphens/underscores
func validateTag(fl validator.FieldLevel) bool {
	tag := fl.Field().String()
	if tag == "" {
		return true // Empty tags in arrays are filtered elsewhere
	}
	return tagRegex.MatchString(tag)
}

// validateTableStyle validates table style values
func validateTableStyle(fl validator.FieldLevel) bool {
	style := fl.Field().String()
	validStyles := map[string]bool{"simple": true, "rounded": true, "unicode": true}
	return validStyles[style]
}

// validateDateFormat validates date format values
func validateDateFormat(fl validator.FieldLevel) bool {
	format := fl.Field().String()
	validFormats := map[string]bool{"rfc3339": true, "relative": true, "short": true}
	return validFormats[format]
}

// ValidateTags validates that all tags in a slice are valid
func ValidateTags(tags []string) error {
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		if !tagRegex.MatchString(tag) {
			return validator.ValidationErrors{}
		}
	}
	return nil
}
