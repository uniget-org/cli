package git

import (
	"fmt"
	"regexp"
	"strings"
)

type PlatformChange struct {
	FilePath  string
	FileName  string
	ToolName  string
	Diff      string
	Added     int
	Removed   int
	DiffLines []string
}

type PlatformChanges struct {
	Changes []PlatformChange
}

type Platform interface {
	GetCommitChanges(fromSha string) (PlatformChanges, error)
	GetMergeChanges(id string) (PlatformChanges, error)
}

func NewPlatformChange(FileName string, Diff string) *PlatformChange {
	change := &PlatformChange{
		FilePath: FileName,
		Diff:     Diff,
	}

	filePathParts := strings.Split(change.FilePath, "/")
	change.FileName = filePathParts[len(filePathParts)-1]
	change.ToolName = change.GetToolName()
	change.DiffLines = strings.Split(Diff, "\n")

	for _, line := range change.DiffLines {
		if strings.HasPrefix(line, "+") {
			change.Added++
		} else if strings.HasPrefix(line, "-") {
			change.Removed++
		}
	}

	return change
}

func (gf *PlatformChange) GetToolName() string {
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

func (gf *PlatformChange) FindChangedFieldsInManifest() []string {
	fields := []string{}

	fieldInDiffRegEx, err := regexp.Compile(`^\+([^:]+):`)
	if err != nil {
		panic(fmt.Errorf("unable to compile regex for field extraction from diff: %s", err))
	}

	if strings.HasSuffix(gf.FilePath, "/manifest.yaml") {
		for _, line := range gf.DiffLines {
			if fieldInDiffRegEx.MatchString(line) {
				fieldName := fieldInDiffRegEx.FindStringSubmatch(line)[1]
				fields = append(fields, fieldName)
			}
		}
	}

	return fields
}
