package buildcodegraph

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func (a *Activities) BuildAstRdf(ctx context.Context, state BuildCodeGraphState) (BuildCodeGraphState, error) {
	tmpDir, _ := os.MkdirTemp("", "rdfControlFlow-*")

	files, err := getFileList(state.AstControlFlowFolderPath)
	if err != nil {
		return state, fmt.Errorf("failed to get file list: %w", err)
	}

	// 2. Use default prompt file path if none provided.
	promptFilePath := "prompts/generate_ast_rdf.txt"

	// 3. Read the prompt template from the file.
	promptTemplate, err := ReadFileToString(promptFilePath)
	if err != nil {
		return state, fmt.Errorf("failed to read prompt file: %w", err)
	}

	// 4. Read the current repository RDF graph.
	currRdf, err := ReadFileToString(state.RepoRdfGraph)
	if err != nil {
		return state, fmt.Errorf("failed to read RDF graph file: %w", err)
	}

	// 5. Read the AST control flow files one by one and call the prompt.
	for _, file := range files {
		fullPath := filepath.Join(state.AstControlFlowFolderPath, file)
		content, err := ReadFileToString(fullPath)
		if err != nil {
			return state, fmt.Errorf("failed to read AST control flow file: %w", err)
		}

		// 6. Substitute placeholders in the prompt template.
		prompt := strings.ReplaceAll(promptTemplate, "{{.ProjectRDF}}", currRdf)
		prompt = strings.ReplaceAll(prompt, "{{.ApiControlFlow}}", content)

		// 7. Call LLM to generate RDF.
		response, err := CallClaudeApi(ctx, prompt)
		if err != nil {
			return state, fmt.Errorf("LLM call failed: %w", err)
		}

		// write rdf to a file
		apiRdf, err := ExtractTurtleRDF(response)
		if err != nil {
			fmt.Printf("Failed parsinng RDF %s: %v\n", response, err)
			continue // continue with other files
		}

		_, err = WriteStringToFile(apiRdf, tmpDir, "api_*.ttl")
		if err != nil {
			return state, fmt.Errorf("failed to write RDF file: %w", err)
		}
	}

	state.AstControlRdfGraph = tmpDir
	return state, nil
}
