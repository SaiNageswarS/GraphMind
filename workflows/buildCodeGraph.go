package workflows

import (
	"time"

	"github.com/SaiNageswarS/GraphMind/activities/buildcodegraph"
	"go.temporal.io/sdk/workflow"
)

func BuildCodeGraphWorkflow(ctx workflow.Context, state buildcodegraph.BuildCodeGraphState) (buildcodegraph.BuildCodeGraphState, error) {
	// Set activity options.
	activityOpts := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute * 10,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOpts)

	activities := &buildcodegraph.Activities{}

	// 1. DownloadRepo Activity: Clone the repository.
	if state.LocalRepoPath == "" {
		// Run the DownloadRepo activity first.
		err := workflow.ExecuteActivity(ctx, activities.DownloadRepo, state).Get(ctx, &state)
		if err != nil {
			return state, err
		}
	}

	// 2. GenerateRDFGraph Activity: Generate an RDF graph from the repository files.
	if state.RepoRdfGraph == "" {
		err := workflow.ExecuteActivity(ctx, activities.GenerateRDFGraph, state).Get(ctx, &state)

		if err != nil {
			return state, err
		}
	}

	// 3. BuildAstControlFlow Activity: Generate AST control flow files for the repository.
	if state.AstControlFlowFolderPath == "" {
		err := workflow.ExecuteActivity(ctx, activities.BuildAstControlFlow, state).Get(ctx, &state)

		if err != nil {
			return state, err
		}
	}

	return state, nil
}
