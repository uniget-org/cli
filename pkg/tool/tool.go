package tool

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

func (tool *Tool) MatchesName(term string) bool {
	match, err := regexp.MatchString(term, tool.Name)
	return err == nil && match
}

func (tool *Tool) HasTag(term string) bool {
	for _, tag := range tool.Tags {
		if tag == term {
			return true
		}
	}

	return false
}

func (tool *Tool) MatchesTag(term string) bool {
	for _, tag := range tool.Tags {
		match, err := regexp.MatchString(term, tag)
		if err == nil && match {
			return true
		}
	}
	return false
}

func (tool *Tool) HasBuildDependency(term string) bool {
	for _, dep := range tool.BuildDependencies {
		if dep == term {
			return true
		}
	}

	return false
}

func (tool *Tool) HasRuntimeDependency(term string) bool {
	for _, dep := range tool.RuntimeDependencies {
		if dep == term {
			return true
		}
	}

	return false
}

func (tool *Tool) MatchesBuildDependency(term string) bool {
	for _, dep := range tool.BuildDependencies {
		match, err := regexp.MatchString(term, dep)
		if err == nil && match {
			return true
		}
	}
	return false
}

func (tool *Tool) MatchesRuntimeDependency(term string) bool {
	for _, dep := range tool.RuntimeDependencies {
		match, err := regexp.MatchString(term, dep)
		if err == nil && match {
			return true
		}
	}
	return false
}

func replaceVariables(source string, variables []string, values []string) (result string) {
	result = source

	for index := range variables {
		result = strings.Replace(result, variables[index], values[index], -1)
	}

	return
}

func (tool *Tool) ReplaceVariables(target string, arch string, altArch string) {
	log.Tracef("Replacing variables for %s", tool.Name)

	//binary
	tool.Binary = replaceVariables(tool.Binary,
		[]string{"${name}", "${target}"},
		[]string{tool.Name, target},
	)
	if tool.Binary[:1] != "/" {
		tool.Binary = target + "/bin/" + tool.Binary
	}

	//check
	tool.Check = replaceVariables(tool.Check,
		[]string{"${binary}", "${name}", "${target}"},
		[]string{tool.Binary, tool.Name, target},
	)
}

func (tool *Tool) GetBinaryStatus() error {
	_, err := os.Stat(tool.Binary)
	if err == nil {
		log.Debugf("Binary for tool %s is present", tool.Name)
		tool.Status.BinaryPresent = true

	} else if errors.Is(err, os.ErrNotExist) {
		log.Debugf("Binary for tool %s is not present", tool.Name)
		tool.Status.BinaryPresent = false

	} else {
		return fmt.Errorf("unable to check binary status for %s: %s", tool.Name, err)
	}

	return nil
}

func (tool *Tool) GetMarkerFileStatus(markerFileDirectory string) error {
	_, err := os.Stat(fmt.Sprintf("%s/%s/%s", markerFileDirectory, tool.Name, tool.Version))
	if err == nil {
		log.Debugf("Marker file for tool %s is present", tool.Name)
		tool.Status.MarkerFilePresent = true

	} else if errors.Is(err, os.ErrNotExist) {
		log.Debugf("Marker file for tool %s is not present", tool.Name)
		tool.Status.MarkerFilePresent = false

	} else {
		return fmt.Errorf("unable to check marker file status for %s: %s", tool.Name, err)
	}

	return nil
}

func (tool *Tool) CreateMarkerFile(markerFileDirectory string) error {
	log.Tracef("Creating marker file for %s", tool.Name)

	err := os.MkdirAll(fmt.Sprintf("%s/%s", markerFileDirectory, tool.Name), 0755)
	if err != nil {
		return fmt.Errorf("unable to create marker file directory for %s: %s", tool.Name, err)
	}

	_, err = os.Create(fmt.Sprintf("%s/%s/%s", markerFileDirectory, tool.Name, tool.Version))
	if err != nil {
		return fmt.Errorf("unable to create marker file for %s: %s", tool.Name, err)
	}

	return nil
}

func (tool *Tool) RemoveMarkerFile(markerFileDirectory string) error {
	log.Tracef("Removing marker file for %s", tool.Name)

	err := os.Remove(fmt.Sprintf("%s/%s/%s", markerFileDirectory, tool.Name, tool.Version))
	if err != nil {
		return fmt.Errorf("unable to remove marker file for %s: %s", tool.Name, err)
	}

	return nil
}

func (tool *Tool) GetVersionStatus() error {
	if tool.Status.BinaryPresent && len(tool.Check) > 0 {
		log.Tracef("Running version check for %s: %s", tool.Name, tool.Check)
		cmd := exec.Command("/bin/bash", "-c", tool.Check+" | tr -d '\n'")
		version, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("unable to execute version check (%s): %s", tool.Check, err)
		}
		tool.Status.Version = string(version)
	}

	log.Tracef("Comparing requested version <%s> with installed version <%s>.", tool.Version, tool.Status.Version)
	tool.Status.VersionMatches = tool.Status.Version == tool.Version

	return nil
}
