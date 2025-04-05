package main

import (
	"log"
	"os"
	"strings"

	"github.com/SaiNageswarS/GraphMind/activities/buildcodegraph"
	"github.com/SaiNageswarS/GraphMind/services"
	"github.com/SaiNageswarS/GraphMind/workflows"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	loadEnv()

	// Create Temporal client.
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	// Create worker for the task queue.
	w := worker.New(c, "GraphMind", worker.Options{})

	// Register the workflow and activities.
	w.RegisterWorkflow(workflows.BuildCodeGraphWorkflow)
	w.RegisterWorkflow(workflows.BuildMultipleCodeGraphsWorkflow)
	w.RegisterActivity(&buildcodegraph.Activities{})

	// Start the HTTP server in a separate goroutine.
	go services.StartHTTPServer()

	// Start listening to the Task Queue.
	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}

func loadEnv(envPath ...string) error {
	if len(envPath) == 0 {
		_, err := os.Stat(".env")

		if err == nil {
			envPath = append(envPath, ".env")
		}
	}

	for _, filename := range envPath {
		content, err := os.ReadFile(filename)
		if err != nil {
			return err
		}

		err = loadEnvFromString(string(content))
		if err != nil {
			return err
		}
	}

	return nil
}

func loadEnvFromString(env string) error {
	lines := strings.Split(env, "\n")
	for _, line := range lines {
		// skip comments
		if strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Split(line, "=")
		key := strings.TrimSpace(parts[0])

		// join rest of the parts with "="
		value := strings.TrimSpace(strings.Join(parts[1:], "="))

		os.Setenv(key, value)
	}

	return nil
}
