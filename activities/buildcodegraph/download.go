package buildcodegraph

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// DownloadRepoInput contains the Git repo URL to clone
type DownloadRepoInput struct {
	RepoURL string // e.g., "https://github.com/your-org/your-repo.git"
}

// DownloadRepoOutput contains the path where the repo was cloned
type DownloadRepoOutput struct {
	LocalPath string
}

// Activities defines all build_code_graph activities
type Activities struct{}

// DownloadRepo clones a Git repository (with submodules) into a temp dir
func (a *Activities) DownloadRepo(ctx context.Context, input DownloadRepoInput) (DownloadRepoOutput, error) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "repo-*")
	if err != nil {
		return DownloadRepoOutput{}, fmt.Errorf("failed to create temp dir: %w", err)
	}

	targetDir := filepath.Join(tmpDir, "repo")

	// Clone the repository with submodules
	cmd := exec.CommandContext(ctx, "git", "clone", "--recurse-submodules", input.RepoURL, targetDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("Cloning repo: %s into %s\n", input.RepoURL, targetDir)

	if err := cmd.Run(); err != nil {
		return DownloadRepoOutput{}, fmt.Errorf("git clone failed: %w", err)
	}

	return DownloadRepoOutput{
		LocalPath: targetDir,
	}, nil
}
