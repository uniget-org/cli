package git

import (
	"fmt"

	gitlab "gitlab.com/gitlab-org/api/client-go"
	"gitlab.com/uniget-org/cli/pkg/logging"
)

type GitLabGitForge struct {
	owner      string
	repository string
	token      string
	client     *gitlab.Client
}

type GitLabGitForgeOption func(*GitLabGitForge)

func NewGitLabGitForge(owner string, repository string, options ...GitLabGitForgeOption) (*GitLabGitForge, error) {
	gitLabGitForge := &GitLabGitForge{}

	for _, opt := range options {
		opt(gitLabGitForge)
	}

	var err error
	gitLabGitForge.client, err = gitlab.NewClient(gitLabGitForge.token)
	if err != nil {
		return gitLabGitForge, fmt.Errorf("error creating client: %s", err)
	}

	return gitLabGitForge, nil
}

func WithGitLabToken(token string) GitLabGitForgeOption {
	return func(GitLabGitForge *GitLabGitForge) {
		GitLabGitForge.token = token
	}
}

func (gl *GitLabGitForge) GetCommitChanges(fromSha string) (GitForgeChanges, error) {
	changes := GitForgeChanges{}

	project, _, err := gl.client.Projects.GetProject(
		"uniget-org/tools",
		&gitlab.GetProjectOptions{},
	)
	if err != nil {
		return changes, fmt.Errorf("unable to find project: %s", err)
	}
	logging.Info.Printfln("Project ID is %d", project.ID)

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
		changes.Changes = append(changes.Changes, GitForgeChange{
			FileName: file.NewPath,
			Diff:     file.Diff,
		})
	}

	return changes, nil
}

func (gl *GitLabGitForge) GetMergeChanges(id string) (GitForgeChanges, error) {
	return GitForgeChanges{}, nil
}
