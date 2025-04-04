package buildcodegraph

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

// GenerateRDFGraphInput holds the parameters for the activity.
type GenerateRDFGraphInput struct {
	FolderPath string `json:"folderPath"` // List of file paths
	RepoURL    string `json:"repoUrl"`    // URL of the git repository
}

// GenerateRDFGraphOutput holds the RDF graph returned from OpenAI.
type GenerateRDFGraphOutput struct {
	RDFGraph string `json:"rdfGraph"`
}

// GenerateRDFGraph reads a prompt template from a file, substitutes the file list and repository URL,
// and calls OpenAI to generate an RDF graph of the repository based solely on its files.
func (a *Activities) GenerateRDFGraph(ctx context.Context, input GenerateRDFGraphInput) (GenerateRDFGraphOutput, error) {
	// Get a list of files recursively from the provided folder path.
	files, err := getFileList(input.FolderPath)
	if err != nil {
		return GenerateRDFGraphOutput{}, fmt.Errorf("failed to get file list: %w", err)
	}

	// Use default prompt file path if none provided.
	promptFilePath := "prompts/generate_repo_metadata.txt"

	// Read the prompt template from the file.
	promptTemplate := ReadFileToString(promptFilePath)

	// Join the file list into a newline-separated string.
	fileListStr := strings.Join(files, "\n")

	// Substitute placeholders in the template.
	// The template should have {{.FileList}} and {{.RepoURL}} as placeholders.
	prompt := strings.ReplaceAll(promptTemplate, "{{.FileList}}", fileListStr)
	prompt = strings.ReplaceAll(prompt, "{{.RepoURL}}", input.RepoURL)

	// Call OpenAI API using the generated prompt.
	response, err := CallOpenAI(ctx, prompt)
	if err != nil {
		return GenerateRDFGraphOutput{}, fmt.Errorf("OpenAI call failed: %w", err)
	}

	rdfPath := WriteRdf(response)
	return GenerateRDFGraphOutput{
		RDFGraph: rdfPath,
	}, nil
}

func getFileList(folderPath string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(folderPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Only include files (skip directories)
		if !d.IsDir() {
			relativePath, err := filepath.Rel(folderPath, path)
			if err != nil {
				// fallback to full path if relative conversion fails
				relativePath = path
			}
			files = append(files, relativePath)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}
