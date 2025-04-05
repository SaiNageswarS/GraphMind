package buildcodegraph

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
)

// callOpenAI calls the OpenAI API to generate text based on the provided prompt.
func CallOpenAI(ctx context.Context, prompt string) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}

	// Prepare the request payload using the chat completions API with GPT-4o.
	payload := map[string]interface{}{
		"model": "gpt-4o",
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"max_tokens": 500,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call OpenAI API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("OpenAI API error: %s", string(bodyBytes))
	}

	// Parse the response from the chat completions endpoint.
	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", fmt.Errorf("failed to decode OpenAI response: %w", err)
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no choices returned from OpenAI")
	}

	return strings.TrimSpace(result.Choices[0].Message.Content), nil
}

func ReadFileToString(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error reading file %s: %v\n", filePath, err)
		return "", err
	}
	return string(data), nil
}

func WriteRdf(content string) (string, error) {
	tempFile, err := os.CreateTemp("", "output_*_rdf")
	if err != nil {
		fmt.Printf("Error creating temp file: %v\n", err)
		return "", err
	}
	// Ensure the file is closed and cleaned up as needed.
	defer tempFile.Close()

	// Write the content to the temporary file.
	if _, err := tempFile.WriteString(content); err != nil {
		fmt.Printf("Error writing to temp file: %v\n", err)
		return "", err
	}

	// Return the full path to the temporary file.
	return tempFile.Name(), nil
}

// extractTurtleRDF extracts the Turtle RDF content from the given text.
func ExtractTurtleRDF(text string) (string, error) {
	// (?s) enables dotall mode so that '.' matches newline characters.
	// The regex captures text between ```turtle and the closing ``` markers.
	re := regexp.MustCompile("(?s)```turtle\\s*(.*?)\\s*```")
	matches := re.FindStringSubmatch(text)
	if len(matches) < 2 {
		return "", fmt.Errorf("no turtle RDF found")
	}
	return matches[1], nil
}
