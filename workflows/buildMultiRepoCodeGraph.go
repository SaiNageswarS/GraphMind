package workflows

import (
	"time"

	"github.com/SaiNageswarS/GraphMind/activities/buildcodegraph"
	"go.temporal.io/sdk/workflow"
)

type BuildMultipleCodeGraphsWorkflowInput struct {
	RepoURLs     []string // Array of repository URLs to process.
	CommonFolder string   // Common folder to store the generated files.
}

// BuildMultipleCodeGraphsWorkflow takes an array of repo URLs and a common folder (temp folder in this case).
// It launches the BuildCodeGraphWorkflow as a child workflow for each repo URL and then copies all the generated
// AstControlRdfGraph files into the common folder.
func BuildMultipleCodeGraphsWorkflow(ctx workflow.Context, input BuildMultipleCodeGraphsWorkflowInput) (string, error) {
	// Set child workflow options.
	childWorkflowOpts := workflow.ChildWorkflowOptions{
		WorkflowRunTimeout: time.Minute * 15,
	}
	ctx = workflow.WithChildOptions(ctx, childWorkflowOpts)

	// Launch a child workflow for each repo URL.
	var childFutures []workflow.Future
	for _, repoURL := range input.RepoURLs {
		// Prepare the initial state for each repository.
		state := buildcodegraph.BuildCodeGraphState{
			RepoURL: repoURL,
		}
		future := workflow.ExecuteChildWorkflow(ctx, BuildCodeGraphWorkflow, state)
		childFutures = append(childFutures, future)
	}

	// Collect the results from all child workflows.
	var results []buildcodegraph.BuildCodeGraphState
	for _, future := range childFutures {
		var result buildcodegraph.BuildCodeGraphState
		if err := future.Get(ctx, &result); err != nil {
			return "", err
		}
		results = append(results, result)
	}

	// Set activity options for the copying activity.
	activityOpts := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute * 10,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOpts)

	activities := &buildcodegraph.Activities{}

	// Call the CopyAstControlRdfGraphs activity with the collected results and the common (temp) folder.
	var combinedRdfFilePath string
	err := workflow.ExecuteActivity(ctx, activities.CopyAstControlRdfGraphs, results, input.CommonFolder).Get(ctx, &combinedRdfFilePath)
	if err != nil {
		return "", err
	}

	return combinedRdfFilePath, nil
}
