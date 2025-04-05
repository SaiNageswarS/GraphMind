package buildcodegraph

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
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

func WriteStringToFile(content, filePath, pattern string) (string, error) {
	tempFile, err := os.CreateTemp(filePath, pattern)
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

func CallClaudeApi(ctx context.Context, prompt string) (string, error) {
	apiKey := os.Getenv("CLAUDE_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("CLAUDE_API_KEY environment variable not set")
	}

	// Prepare the request payload for Claude API.
	payload := map[string]interface{}{
		"model": "claude-3-5-sonnet-20241022",
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"max_tokens": 2048,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call Claude API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("claude api error: %s", string(bodyBytes))
	}

	var result struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil || len(result.Content) == 0 {
		return "", fmt.Errorf("failed to decode Claude response: %w", err)
	}

	return strings.TrimSpace(result.Content[0].Text), nil
}

func CallUnifyRdfsApi(rdfsBasePath, outputFile string) (string, error) {
	// Create the JSON payload.
	payload := struct {
		Folder string `json:"folder"`
		Output string `json:"output"`
	}{
		Folder: rdfsBasePath,
		Output: outputFile,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Define the Python API endpoint.
	url := "http://localhost:5000/unify"

	// Create the POST request.
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Execute the request.
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error making request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response.
	var result struct {
		CombinedGraphPath string `json:"combined_graph_path"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.CombinedGraphPath, nil
}
