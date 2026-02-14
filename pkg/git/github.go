package git

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-github/v81/github"
	"gitlab.com/uniget-org/cli/pkg/logging"
)

type GitHubPlatform struct {
	owner        string
	repository   string
	token        string
	client       *github.Client
	registryHost string
}

type GitHubPlatformOption func(*GitHubPlatform)

func NewGitHubPlatform(owner string, repository string, options ...GitHubPlatformOption) *GitHubPlatform {
	gitHubPlatform := &GitHubPlatform{
		owner:        owner,
		repository:   repository,
		registryHost: "ghcr.io",
	}

	for _, opt := range options {
		opt(gitHubPlatform)
	}

	gitHubPlatform.client = github.NewClient(nil)
	if gitHubPlatform.token != "" {
		gitHubPlatform.client = gitHubPlatform.client.WithAuthToken(gitHubPlatform.token)
	}

	return gitHubPlatform
}

func WithGitHubToken(token string) GitHubPlatformOption {
	return func(gitHubPlatform *GitHubPlatform) {
		gitHubPlatform.token = token
	}
}

func WithGitHubTokenFromEnv() GitHubPlatformOption {
	return func(gitHubPlatform *GitHubPlatform) {
		gitHubPlatform.token = os.Getenv("GITHUB_TOKEN")
	}
}

func (gh *GitHubPlatform) GetRepositoryPath() (string, error) {
	return fmt.Sprintf("%s/%s", gh.owner, gh.repository), nil
}

func (gh *GitHubPlatform) GetRegistryHost() (string, error) {
	return gh.registryHost, nil
}

func (gh *GitHubPlatform) GetCommitChanges(fromSha string) (PlatformChanges, error) {
	changes := PlatformChanges{}

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
		changes.Changes = append(changes.Changes, *NewPlatformChange(
			file.GetFilename(),
			file.GetPatch(),
		))
	}

	return changes, nil
}

func (gh *GitHubPlatform) GetMergeChanges(id string) (PlatformChanges, error) {
	return PlatformChanges{}, nil
}
