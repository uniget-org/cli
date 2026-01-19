package git

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-github/v81/github"
	"gitlab.com/uniget-org/cli/pkg/logging"
)

type GitHubGitForge struct {
	owner      string
	repository string
	token      string
	client     *github.Client
}

type GitHubGitForgeOption func(*GitHubGitForge)

func NewGitHubGitForge(owner string, repository string, options ...GitHubGitForgeOption) *GitHubGitForge {
	gitHubGitForge := &GitHubGitForge{
		owner:      owner,
		repository: repository,
	}

	for _, opt := range options {
		opt(gitHubGitForge)
	}

	gitHubGitForge.client = github.NewClient(nil)
	if gitHubGitForge.token != "" {
		gitHubGitForge.client = gitHubGitForge.client.WithAuthToken(gitHubGitForge.token)
	}

	return gitHubGitForge
}

func WithGitHubToken(token string) GitHubGitForgeOption {
	return func(gitHubGitForge *GitHubGitForge) {
		gitHubGitForge.token = token
	}
}

func WithGitHubTokenFromEnv() GitHubGitForgeOption {
	return func(gitHubGitForge *GitHubGitForge) {
		gitHubGitForge.token = os.Getenv("GITHUB_TOKEN")
	}
}

func (gh *GitHubGitForge) GetCommitChanges(fromSha string) (GitForgeChanges, error) {
	changes := GitForgeChanges{}

	repo, _, err := gh.client.Repositories.Get(
		context.Background(),
		gh.owner,
		gh.repository,
	)
	if err != nil {
		return changes, fmt.Errorf("unable to get repository uniget-org/tools: %s", err)
	}
	logging.Debugf("GitHub default branch: %s", repo.GetDefaultBranch())

	fromShaCommit, _, err := gh.client.Repositories.GetCommit(
		context.Background(),
		gh.owner,
		gh.repository,
		fromSha,
		&github.ListOptions{},
	)
	if err != nil {
		return changes, fmt.Errorf("unable to get source commit: %s", err)
	}

	headShaCommit, _, err := gh.client.Repositories.GetCommit(
		context.Background(),
		gh.owner,
		gh.repository,
		repo.GetDefaultBranch(),
		&github.ListOptions{},
	)
	if err != nil {
		return changes, fmt.Errorf("unable to get head commit: %s", err)
	}

	comparison, _, err := gh.client.Repositories.CompareCommits(
		context.Background(),
		gh.owner,
		gh.repository,
		fromShaCommit.GetSHA(),
		headShaCommit.GetSHA(),
		&github.ListOptions{},
	)
	if err != nil {
		return changes, fmt.Errorf("unable to compare commits: %s", err)
	}

	for _, file := range comparison.Files {
		changes.Changes = append(changes.Changes, *NewGitForgeChange(
			file.GetFilename(),
			file.GetPatch(),
		))
	}

	return changes, nil
}

func (gh *GitHubGitForge) GetMergeChanges(id string) (GitForgeChanges, error) {
	return GitForgeChanges{}, nil
}
