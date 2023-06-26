package tool

import (
	"fmt"
	"regexp"

	log "github.com/sirupsen/logrus"
)

func (tools *Tools) Contains(name string) bool {
	for _, tool := range tools.Tools {
		if tool.Name == name {
			return true
		}
	}
	return false
}

func (tools *Tools) GetByName(name string) (*Tool, error) {
	for _, tool := range tools.Tools {
		if tool.Name == name {
			return &tool, nil
		}
	}

	return nil, fmt.Errorf("Tool named %s not found", name)
}

func (tools *Tools) GetByTag(tagName string) *Tools {
	var toolList Tools

	for _, tool := range tools.Tools {
		if tool.HasTag(tagName) {
			toolList.Tools = append(toolList.Tools, tool)
		}
	}

	return &toolList
}

func (tools *Tools) GetByNames(names []string) Tools {
	var toolList Tools

	for _, tool := range tools.Tools {
		for _, name := range names {
			if tool.Name == name {
				toolList.Tools = append(toolList.Tools, tool)
			}
		}
	}

	return toolList
}

func (tools *Tools) GetByTags(tagNames []string) Tools {
	var toolList Tools

	for _, tool := range tools.Tools {
		for _, tag := range tagNames {
			if tool.HasTag(tag) {
				toolList.Tools = append(toolList.Tools, tool)
			}
		}
	}

	return toolList
}

func (tools *Tools) Find(term string, searchInName bool, searchInTags bool, searchInDeps bool) Tools {
	var results = Tools{}

	for _, tool := range tools.Tools {
		matches := false

		if searchInName && tool.MatchesName(term) {
			matches = true
		}

		for _, tag := range tool.Tags {
			match, err := regexp.MatchString(term, tag)
			if err == nil && searchInTags && match {
				matches = true
			}
		}

		for _, dep := range tool.RuntimeDependencies {
			match, err := regexp.MatchString(term, dep)
			if err == nil && searchInDeps && match {
				matches = true
			}
		}

		if matches {
			results.Tools = append(results.Tools, tool)
		}
	}

	return results
}

func (tools *Tools) GetNames() []string {
	var toolNames []string

	for _, tool := range tools.Tools {
		toolNames = append(toolNames, tool.Name)
	}

	return toolNames
}

func (tools *Tools) AddIfMissing(newTool *Tool) {
	for _, tool := range tools.Tools {
		if tool.Name == newTool.Name {
			return
		}
	}

	tools.Tools = append(tools.Tools, *newTool)
}

func (tools *Tools) ResolveDependencies(queue *Tools, toolName string) error {
	log.Tracef("Resolving dependencies for %s", toolName)

	tool, err := tools.GetByName(toolName)
	if err != nil {
		log.Errorf("Error resolving dependencies for %s", toolName)
		return err
	}

	for _, depName := range tool.RuntimeDependencies {
		log.Tracef("Recursing for dependency %s", depName)

		dep, err := tools.GetByName(depName)
		if err != nil {
			log.Errorf("Unable to find dependency called %s for %s", depName, toolName)
		}

		err = tools.ResolveDependencies(queue, depName)
		if err != nil {
			return err
		}

		dep.Status.IsDependency = true
		queue.AddIfMissing(dep)
	}

	queue.AddIfMissing(tool)

	return nil
}
