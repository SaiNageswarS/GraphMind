package buildcodegraph

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

// GenerateRDFGraph reads a prompt template from a file, substitutes the file list and repository URL,
// and calls OpenAI to generate an RDF graph of the repository based solely on its files.
func (a *Activities) GenerateRDFGraph(ctx context.Context, state BuildCodeGraphState) (BuildCodeGraphState, error) {
	// 1. Retrieve a list of files from the provided folder path.
	files, err := getFileList(state.LocalRepoPath)
	if err != nil {
		return state, fmt.Errorf("failed to get file list: %w", err)
	}

	// 2. Use default prompt file path if none provided.
	promptFilePath := "prompts/generate_repo_metadata.txt"

	// 3. Read the prompt template from the file.
	promptTemplate, err := ReadFileToString(promptFilePath)
	if err != nil {
		return state, fmt.Errorf("failed to read prompt file: %w", err)
	}

	// 4. Build the file list string.
	fileListStr := strings.Join(files, "\n")

	// 5. Look for language-specific configuration files and combine their contents.
	configFiles := []string{"go.mod", "build.gradle", "packages.json", "requirements.txt"}
	var additionalInfoBuilder strings.Builder
	for _, fileName := range configFiles {
		if containsFile(files, fileName) {
			fullPath := filepath.Join(state.LocalRepoPath, fileName)
			content, err := ReadFileToString(fullPath)
			if err != nil {
				additionalInfoBuilder.WriteString(fmt.Sprintf("Error reading %s\n", fileName))
			} else {
				additionalInfoBuilder.WriteString(fmt.Sprintf("== %s ==\n%s\n", fileName, content))
			}
		}
	}
	additionalInfo := additionalInfoBuilder.String()
	if additionalInfo == "" {
		additionalInfo = "None"
	}

	// 6. Substitute placeholders in the prompt template.
	prompt := strings.ReplaceAll(promptTemplate, "{{.FileList}}", fileListStr)
	prompt = strings.ReplaceAll(prompt, "{{.AdditionalInfo}}", additionalInfo)
	prompt = strings.ReplaceAll(prompt, "{{.RepoURL}}", state.RepoURL)

	// 7. Call OpenAI API using GPT-4o to generate RDF.
	response, err := CallOpenAI(ctx, prompt)
	if err != nil {
		return state, fmt.Errorf("OpenAI call failed: %w", err)
	}

	response, err = ExtractTurtleRDF(response)
	if err != nil {
		return state, fmt.Errorf("failed to extract Turtle RDF: %w", err)
	}

	// 8. Write the RDF content to a file.
	rdfPath, err := WriteStringToFile(response, "", "repo_metadata_*.ttl")
	if err != nil {
		return state, fmt.Errorf("failed to write RDF file: %w", err)
	}

	state.RepoRdfGraph = rdfPath
	return state, nil
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

func containsFile(files []string, filename string) bool {
	for _, f := range files {
		if filepath.Base(f) == filename {
			return true
		}
	}
	return false
}
