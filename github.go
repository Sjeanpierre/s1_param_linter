package main

import (
	"github.com/google/go-github/github"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"log"
	"os"
)

var StatusState = map[bool]string{true: "success", false: "failure"}
var StatusDescription = map[bool]string{true: "Looks good, one service updated", false: "Multiple services modified :-("}

func githubClient() (client *github.Client) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_API_KEY")},
	)
	tc := oauth2.NewClient(ctx, ts)

	client = github.NewClient(tc)
	return
}

func PostStatus(commits []CommitSet) {
	ctx := context.Background()
	ghClient := githubClient()
	for _, commit := range commits {
		stat := github.RepoStatus{
			State:       stringPointer(StatusState[commit.Clean]),
			Description: stringPointer(StatusDescription[commit.Clean]),
			Context:     stringPointer("status-bot/s1_params"),
			TargetURL:   stringPointer("https://gist.github.com/Sjeanpierre/09e29f41ef4b8b49a81060121906747c"),
		}

		status, response, err := ghClient.Repositories.CreateStatus(
			ctx,
			commit.RepoOwner,
			commit.RepoName,
			commit.CommitID,
			&stat,
		)
		if err != nil {
			log.Printf("Encountered error posting status for %+v\n %s\n Response: %s\n", commit, err, response)
			continue
		}
		log.Printf("Status posted %s", status)
	}
}

func stringPointer(x string) *string {
	return &x
}
