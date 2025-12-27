package source

import (
	"fmt"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

// CloneRepository clones a Git repository to the specified local path.
// If a token is provided, it uses authenticated requests.
// Returns the current commit SHA and any error.
func CloneRepository(url, localPath, token string) (commitSHA string, err error) {
	// Prepare clone options
	cloneOpts := &git.CloneOptions{
		URL:      url,
		Progress: os.Stderr, // Show progress to stderr
	}

	// Add authentication if token provided
	if token != "" {
		cloneOpts.Auth = &http.BasicAuth{
			Username: "x-access-token", // GitHub uses any username with token
			Password: token,
		}
	}

	// Clone repository
	repo, err := git.PlainClone(localPath, false, cloneOpts)
	if err != nil {
		return "", fmt.Errorf("failed to clone repository: %w", err)
	}

	// Get HEAD commit SHA
	ref, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD reference: %w", err)
	}

	return ref.Hash().String(), nil
}

// PullRepository updates an existing Git repository to the latest commit.
// If a token is provided, it uses authenticated requests.
// Returns the new commit SHA and any error.
func PullRepository(localPath, token string) (commitSHA string, err error) {
	// Open existing repository
	repo, err := git.PlainOpen(localPath)
	if err != nil {
		return "", fmt.Errorf("failed to open repository: %w", err)
	}

	// Get working tree
	worktree, err := repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %w", err)
	}

	// Prepare pull options
	pullOpts := &git.PullOptions{
		Progress: os.Stderr, // Show progress to stderr
	}

	// Add authentication if token provided
	if token != "" {
		pullOpts.Auth = &http.BasicAuth{
			Username: "x-access-token",
			Password: token,
		}
	}

	// Pull latest changes
	err = worktree.Pull(pullOpts)
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return "", fmt.Errorf("failed to pull repository: %w", err)
	}

	// Get HEAD commit SHA
	ref, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD reference: %w", err)
	}

	return ref.Hash().String(), nil
}

// GetCurrentCommitSHA returns the current commit SHA of a local repository.
func GetCurrentCommitSHA(localPath string) (string, error) {
	// Open repository
	repo, err := git.PlainOpen(localPath)
	if err != nil {
		return "", fmt.Errorf("failed to open repository: %w", err)
	}

	// Get HEAD reference
	ref, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD reference: %w", err)
	}

	return ref.Hash().String(), nil
}

// FetchRemote fetches remote changes without merging (for checking updates).
// Returns the remote HEAD commit SHA and any error.
func FetchRemote(localPath, token string) (remoteSHA string, err error) {
	// Open repository
	repo, err := git.PlainOpen(localPath)
	if err != nil {
		return "", fmt.Errorf("failed to open repository: %w", err)
	}

	// Prepare fetch options
	fetchOpts := &git.FetchOptions{
		Progress: nil, // Silent fetch for update checks
	}

	// Add authentication if token provided
	if token != "" {
		fetchOpts.Auth = &http.BasicAuth{
			Username: "x-access-token",
			Password: token,
		}
	}

	// Fetch from remote
	err = repo.Fetch(fetchOpts)
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return "", fmt.Errorf("failed to fetch from remote: %w", err)
	}

	// Get remote HEAD reference
	remote, err := repo.Remote("origin")
	if err != nil {
		return "", fmt.Errorf("failed to get remote: %w", err)
	}

	refs, err := remote.List(&git.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to list remote refs: %w", err)
	}

	// Find HEAD ref
	for _, ref := range refs {
		if ref.Name() == plumbing.HEAD {
			return ref.Hash().String(), nil
		}
	}

	// Fallback: get refs/heads/main or refs/heads/master
	for _, ref := range refs {
		if ref.Name() == plumbing.NewBranchReferenceName("main") ||
			ref.Name() == plumbing.NewBranchReferenceName("master") {
			return ref.Hash().String(), nil
		}
	}

	return "", fmt.Errorf("could not determine remote HEAD")
}

// CheckForUpdates checks if the local repository is behind the remote.
// Returns true if updates are available, along with the remote SHA.
func CheckForUpdates(localPath, token string) (hasUpdates bool, remoteSHA string, err error) {
	// Get current local commit
	localSHA, err := GetCurrentCommitSHA(localPath)
	if err != nil {
		return false, "", err
	}

	// Fetch remote changes
	remoteSHA, err = FetchRemote(localPath, token)
	if err != nil {
		return false, "", err
	}

	// Compare SHAs
	hasUpdates = localSHA != remoteSHA

	return hasUpdates, remoteSHA, nil
}

// GetRepositoryStatus returns a human-readable status of the repository.
// Returns status information including current commit and clean/dirty state.
func GetRepositoryStatus(localPath string) (string, error) {
	// Open repository
	repo, err := git.PlainOpen(localPath)
	if err != nil {
		return "", fmt.Errorf("failed to open repository: %w", err)
	}

	// Get working tree
	worktree, err := repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %w", err)
	}

	// Get status
	status, err := worktree.Status()
	if err != nil {
		return "", fmt.Errorf("failed to get status: %w", err)
	}

	// Check if clean
	if status.IsClean() {
		return "clean", nil
	}

	return "dirty", nil
}
