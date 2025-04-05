package buildcodegraph

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// DownloadRepoInput contains the Git repo URL to clone
type BuildCodeGraphState struct {
	RepoURL                  string // The Git repository URL to clone.
	LocalRepoPath            string // The local path to the cloned repository.
	RepoRdfGraph             string // The RDF graph generated from the repository files.
	AstControlFlowFolderPath string // The path to the folder containing AST control flow files.
	AstControlRdfGraph       string // The RDF graph generated from the AST control flow files.
}

// Activities defines all build_code_graph activities
type Activities struct{}

// DownloadRepo clones a Git repository (with submodules) into a temp dir
func (a *Activities) DownloadRepo(ctx context.Context, state BuildCodeGraphState) (BuildCodeGraphState, error) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "repo-*")
	if err != nil {
		return state, fmt.Errorf("failed to create temp dir: %w", err)
	}

	targetDir := filepath.Join(tmpDir, "repo")

	// Clone the repository with submodules
	cmd := exec.CommandContext(ctx, "git", "clone", "--recurse-submodules", state.RepoURL, targetDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("Cloning repo: %s into %s\n", state.RepoURL, targetDir)

	if err := cmd.Run(); err != nil {
		return state, fmt.Errorf("git clone failed: %w", err)
	}

	state.LocalRepoPath = targetDir
	return state, nil
}
