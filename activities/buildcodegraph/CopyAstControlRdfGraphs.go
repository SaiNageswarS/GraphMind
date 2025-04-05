package buildcodegraph

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func (a *Activities) CopyAstControlRdfGraphs(results []BuildCodeGraphState, commonFolder string) error {
	// Ensure the commonFolder exists.
	if err := os.MkdirAll(commonFolder, 0755); err != nil {
		return err
	}

	// Iterate over each state's result.
	for repoIndex, state := range results {
		// Skip if there is no valid AstControlRdfGraph folder.
		if state.AstControlRdfGraph == "" {
			continue
		}

		// Read all files in the AstControlRdfGraph folder.
		files, err := os.ReadDir(state.AstControlRdfGraph)
		if err != nil {
			return fmt.Errorf("failed to read folder %s: %w", state.AstControlRdfGraph, err)
		}

		// Iterate over each file and copy it.
		for _, file := range files {
			// Skip directories.
			if file.IsDir() {
				continue
			}

			srcFilePath := filepath.Join(state.AstControlRdfGraph, file.Name())
			// Create a unique filename using the repo index and original file name.
			dstFileName := fmt.Sprintf("repo%d_%s", repoIndex, file.Name())
			dstFilePath := filepath.Join(commonFolder, dstFileName)

			// Copy the file from the source to the destination.
			if err := copyFile(srcFilePath, dstFilePath); err != nil {
				return fmt.Errorf("failed to copy file %s to %s: %w", srcFilePath, dstFilePath, err)
			}
		}
	}

	// Combine all RDF files into a single file.
	combinedRdfFilePath := filepath.Join(commonFolder, "combined_rdf.ttl")
	if _, err := CallUnifyRdfsApi(commonFolder, combinedRdfFilePath); err != nil {
		return fmt.Errorf("failed to combine RDF files: %w", err)
	}

	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		_ = out.Close()
	}()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	// Ensure the copied file is flushed to disk.
	return out.Sync()
}
