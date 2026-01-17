package git

import (
	"fmt"
	"iter"
	"regexp"
	"strings"
)

type GitForgeChange struct {
	FilePath  string
	FileName  string
	ToolName  string
	Diff      string
	Added     int
	Removed   int
	DiffLines iter.Seq[string]
}

type GitForgeChanges struct {
	Changes []GitForgeChange
}

type GitForge interface {
	GetCommitChanges(fromSha string) (GitForgeChanges, error)
	GetMergeChanges(id string) (GitForgeChanges, error)
}

func NewGitForgeChange(FileName string, Diff string) *GitForgeChange {
	change := &GitForgeChange{
		FilePath: FileName,
		Diff:     Diff,
	}

	filePathParts := strings.Split(change.FilePath, "/")
	change.FileName = filePathParts[len(filePathParts)-1]
	change.ToolName = change.GetToolName()
	change.DiffLines = strings.SplitSeq(Diff, "\n")

	for line := range change.DiffLines {
		if strings.HasPrefix(line, "+") {
			change.Added++
		} else if strings.HasPrefix(line, "-") {
			change.Removed++
		}
	}

	return change
}

func (gf *GitForgeChange) GetToolName() string {
	toolName := ""

	toolNameRegEx, err := regexp.Compile(`^tools/([^/]+)/`)
	if err != nil {
		panic(fmt.Errorf("unable to compile regex for tool name extraction: %s", err))
	}

	if toolNameRegEx.MatchString(gf.FilePath) {
		toolName = toolNameRegEx.FindStringSubmatch(gf.FilePath)[1]
	}

	return toolName
}

func (gf *GitForgeChange) FindChangedFieldsInManifest() []string {
	fields := []string{}

	fieldInDiffRegEx, err := regexp.Compile(`^\+([^:]+):`)
	if err != nil {
		panic(fmt.Errorf("unable to compile regex for field extraction from diff: %s", err))
	}

	if strings.HasSuffix(gf.FilePath, "/manifest.yaml") {
		for line := range gf.DiffLines {
			if fieldInDiffRegEx.MatchString(line) {
				fields = append(fields, fieldInDiffRegEx.FindStringSubmatch(line)[1])
			}
		}
	}

	return fields
}
