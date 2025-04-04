package workflows

import (
	"time"

	"github.com/SaiNageswarS/GraphMind/activities/buildcodegraph"
	"go.temporal.io/sdk/workflow"
)

type BuildCodeGraphWorkflowInput struct {
	StartActivity string // The name of the activity to start the workflow with.
	RepoURL       string // The Git repository URL to clone.
	LocalRepoPath string // The local path to the cloned repository.
	RdfGraph      string // The RDF graph generated from the repository files.
}

// BuildCodeGraphWorkflowOutput defines the output of the workflow.
type BuildCodeGraphWorkflowOutput struct {
	RDFGraph string // The generated RDF graph in Turtle format.
}

func BuildCodeGraphWorkflow(ctx workflow.Context, input BuildCodeGraphWorkflowInput) (BuildCodeGraphWorkflowOutput, error) {
	// Set activity options.
	activityOpts := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute * 10,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOpts)

	startActivity := input.StartActivity
	if startActivity == "" {
		startActivity = "DownloadRepo"
	}

	activities := &buildcodegraph.Activities{}

	// 1. DownloadRepo Activity: Clone the repository.
	if startActivity == "DownloadRepo" {
		// Run the DownloadRepo activity first.
		var downloadOutput buildcodegraph.DownloadRepoOutput
		err := workflow.ExecuteActivity(ctx, activities.DownloadRepo, buildcodegraph.DownloadRepoInput{
			RepoURL: input.RepoURL,
		}).Get(ctx, &downloadOutput)
		if err != nil {
			return BuildCodeGraphWorkflowOutput{}, err
		}
		input.LocalRepoPath = downloadOutput.LocalPath
		startActivity = "GenerateRDFGraph"
	}

	// 2. GenerateRDFGraph Activity: Generate an RDF graph from the repository files.
	if startActivity == "GenerateRDFGraph" {
		var generateOutput buildcodegraph.GenerateRDFGraphOutput
		err := workflow.ExecuteActivity(ctx, activities.GenerateRDFGraph, buildcodegraph.GenerateRDFGraphInput{
			FolderPath: input.LocalRepoPath,
			RepoURL:    input.RepoURL,
		}).Get(ctx, &generateOutput)

		if err != nil {
			return BuildCodeGraphWorkflowOutput{}, err
		}

		input.RdfGraph = generateOutput.RDFGraph
	}

	return BuildCodeGraphWorkflowOutput{
		RDFGraph: input.RdfGraph,
	}, nil
}
