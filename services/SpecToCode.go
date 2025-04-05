package services

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"

	"github.com/SaiNageswarS/GraphMind/activities/buildcodegraph"
)

// startHTTPServer starts an HTTP server that serves the spec input page.
func StartHTTPServer() {
	http.HandleFunc("/", specHandler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port if not set in environment.
	}
	log.Println("HTTP server listening on port " + port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("HTTP server failed: %v", err)
	}
}

// specHandler handles GET and POST requests on the root endpoint.
func specHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		renderTemplate(w, "templates/spec_form.html", nil)
	case http.MethodPost:
		// Parse form data.
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Unable to parse form", http.StatusBadRequest)
			return
		}
		spec := r.FormValue("spec")
		graphFile := r.FormValue("graphLocation")

		// Here, insert your logic to process the spec and the combined RDF graph.
		// For instance, you could call a Temporal workflow or a local function.
		// The following is a dummy output to demonstrate the concept.
		result := processSpec(spec, graphFile)
		data := struct {
			Spec          string
			GraphLocation string
			Result        string
		}{
			Spec:          spec,
			GraphLocation: graphFile,
			Result:        result,
		}
		renderTemplate(w, "templates/spec_form.html", data)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// renderTemplate is a helper to render HTML templates.
func renderTemplate(w http.ResponseWriter, tmplPath string, data interface{}) {
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("Template parsing error: %v", err)
		return
	}
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("Template execution error: %v", err)
	}
}

// processSpec is a placeholder function that simulates processing the spec and RDF.
func processSpec(spec, graphFile string) string {
	promptFilePath := "prompts/spec_to_code.txt"

	promptTemplate, err := buildcodegraph.ReadFileToString(promptFilePath)
	if err != nil {
		log.Printf("Failed to read prompt file: %v", err)
		return "Error reading prompt file."
	}

	graphFileContent, err := buildcodegraph.ReadFileToString(graphFile)
	if err != nil {
		log.Printf("Failed to read graph file: %v", err)
		return "Error reading graph file."
	}

	prompt := strings.ReplaceAll(promptTemplate, "{{.Spec}}", spec)
	prompt = strings.ReplaceAll(prompt, "{{.CombinedRdf}}", graphFileContent)

	response, err := buildcodegraph.CallClaudeApi(context.Background(), prompt)
	if err != nil {
		log.Printf("LLM call failed: %v", err)
		return "LLM call failed."
	}

	log.Printf("LLM response: %s", response)
	return response
}
