package git

import (
	"fmt"
	"os"

	gitlab "gitlab.com/gitlab-org/api/client-go"
	"gitlab.com/uniget-org/cli/pkg/logging"
)

type GitLabPlatform struct {
	owner        string
	repository   string
	token        string
	client       *gitlab.Client
	registryHost string
}

type GitLabPlatformOption func(*GitLabPlatform)

func NewGitLabPlatform(owner string, repository string, options ...GitLabPlatformOption) (*GitLabPlatform, error) {
	gitLabPlatform := &GitLabPlatform{
		owner:        owner,
		repository:   repository,
		registryHost: "registry.gitlab.com",
	}

	for _, opt := range options {
		opt(gitLabPlatform)
	}

	var err error
	gitLabPlatform.client, err = gitlab.NewClient(gitLabPlatform.token)
	if err != nil {
		return gitLabPlatform, fmt.Errorf("error creating client: %s", err)
	}

	return gitLabPlatform, nil
}

func WithGitLabToken(token string) GitLabPlatformOption {
	return func(GitLabPlatform *GitLabPlatform) {
		GitLabPlatform.token = token
	}
}

func WithGitLabJobToken() GitLabPlatformOption {
	return func(GitLabPlatform *GitLabPlatform) {
		GitLabPlatform.token = os.Getenv("CI_JOB_TOKEN")
	}
}

func (gl *GitLabPlatform) GetRepositoryPath() (string, error) {
	return fmt.Sprintf("%s/%s", gl.owner, gl.repository), nil
}

func (gl *GitLabPlatform) GetRegistryHost() (string, error) {
	return gl.registryHost, nil
}

func (gl *GitLabPlatform) GetCommitChanges(fromSha string) (PlatformChanges, error) {
	changes := PlatformChanges{}

	project, _, err := gl.client.Projects.GetProject(
		fmt.Sprintf("%s/%s", gl.owner, gl.repository),
		&gitlab.GetProjectOptions{},
	)
	if err != nil {
		return changes, fmt.Errorf("unable to find project: %s", err)
	}
	logging.Debugf("Project ID is %d", project.ID)

	fromShaCommit, _, err := gl.client.Commits.GetCommit(
		project.ID,
		fromSha,
		&gitlab.GetCommitOptions{},
	)
	if err != nil {
		return changes, fmt.Errorf("failed to get source commit: %s", err)
	}

	toShaCommit, _, err := gl.client.Commits.GetCommit(
		project.ID,
		project.DefaultBranch,
		&gitlab.GetCommitOptions{},
	)
	if err != nil {
		return changes, fmt.Errorf("failed to get source commit: %s", err)
	}

	comparison, _, err := gl.client.Repositories.Compare(
		project.ID,
		&gitlab.CompareOptions{
			From:     &fromShaCommit.ID,
			To:       &toShaCommit.ID,
			Straight: gitlab.Ptr(false),
			Unidiff:  gitlab.Ptr(true),
		},
	)
	if err != nil {
		return changes, fmt.Errorf("failed to compare: %s", err)
	}
	for _, file := range comparison.Diffs {
		changes.Changes = append(changes.Changes, *NewPlatformChange(
			file.NewPath,
			file.Diff,
		))
	}

	return changes, nil
}

func (gl *GitLabPlatform) GetMergeChanges(id string) (PlatformChanges, error) {
	return PlatformChanges{}, nil
}
