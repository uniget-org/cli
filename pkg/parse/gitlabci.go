package parse

import (
	"fmt"
	"reflect"

	"github.com/regclient/regclient/types/ref"
	"gitlab.com/uniget-org/cli/pkg/logging"
	myos "gitlab.com/uniget-org/cli/pkg/os"
	"gitlab.com/uniget-org/cli/pkg/tool"
	"go.yaml.in/yaml/v3"
)

type GitLabCiService struct {
	Name string
}

type GitlabCiDefaults struct {
	Image    string
	Services []GitLabCiService
}

type GitlabCiJob struct {
	Image    string
	Services []GitLabCiService
}

type GitlabCi struct {
	Defaults GitlabCiDefaults
	Jobs     map[string]GitlabCiJob
	Services []GitLabCiService
}

func ParseServices(servicesObject []any) []GitLabCiService {
	jobServices := make([]GitLabCiService, 0)

	for _, serviceObject := range servicesObject {
		switch reflect.TypeOf(serviceObject).String() {
		case "string":
			jobServices = append(jobServices, GitLabCiService{
				Name: fmt.Sprintf("%v", serviceObject),
			})

		case "map[string]interface {}":
			jobServices = append(jobServices, GitLabCiService{
				Name: fmt.Sprintf("%v", serviceObject.(map[string]any)["name"]),
			})
		}
	}

	return jobServices
}

func LoadGitlabCi(file []byte) (GitlabCi, error) {
	var pipelineYaml map[string]any

	err := yaml.Unmarshal(file, &pipelineYaml)
	if err != nil {
		return GitlabCi{}, fmt.Errorf("unable to load GitLab CI from byte array: %s", err)
	}

	pipeline := GitlabCi{
		Jobs: make(map[string]GitlabCiJob, 0),
	}
	for key, value := range pipelineYaml {

		switch key {
		case "include", "workflow":
			// unneeded

		case "default":
			defaults := value.(map[string]any)
			pipeline.Defaults = GitlabCiDefaults{
				Image:    fmt.Sprintf("%v", defaults["image"]),
				Services: make([]GitLabCiService, 0),
			}

			if _, ok := defaults["services"]; ok {
				services := defaults["services"].([]any)
				for _, value := range services {
					service := value.(map[string]any)
					pipeline.Defaults.Services = append(pipeline.Defaults.Services, GitLabCiService{
						Name: fmt.Sprintf("%v", service["name"]),
					})
				}
			}

		default:
			job := value.(map[string]any)

			image := job["image"]
			gitlabCiJob := GitlabCiJob{
				Image:    fmt.Sprintf("%v", image),
				Services: make([]GitLabCiService, 0),
			}

			if _, ok := job["services"]; ok {
				gitlabCiJob.Services = ParseServices(job["services"].([]any))
			}

			pipeline.Jobs[key] = gitlabCiJob
		}
	}

	return pipeline, nil
}

func LoadGitlabCiFromFile(filename string) (GitlabCi, error) {
	file, err := myos.SlurpFile(filename)
	if err != nil {
		return GitlabCi{}, fmt.Errorf("failed to read file: %w", err)
	}

	pipeline, err := LoadGitlabCi(file)
	if err != nil {
		return GitlabCi{}, fmt.Errorf("failed to load GitLab CI: %w", err)
	}

	return pipeline, nil
}

func ExtractImageReferencesFromGitlabCi(pipeline GitlabCi) (ImageRefs, error) {
	var imageRefs ImageRefs

	images := map[string]struct{}{}
	if len(pipeline.Defaults.Image) > 0 {
		images[pipeline.Defaults.Image] = struct{}{}
	}
	for _, service := range pipeline.Defaults.Services {
		if len(service.Name) > 0 {
			images[service.Name] = struct{}{}
		}
	}
	for _, job := range pipeline.Jobs {
		if len(job.Image) > 0 {
			images[job.Image] = struct{}{}
		}
		for _, service := range job.Services {
			if len(service.Name) > 0 {
				images[service.Name] = struct{}{}
			}
		}
	}

	for image := range images {
		imageRef, err := ref.New(image)
		if err != nil {
			logging.Debugf("Failed to create image reference from %s: %v", image, err)
			continue
		}
		imageRefs.Add(imageRef)
	}

	return imageRefs, nil
}

func BumpGitlabCiFile(filename string, tools *tool.Tools) error {
	pipeline, err := LoadGitlabCiFromFile(filename)
	if err != nil {
		return fmt.Errorf("failed to load GitLab CI file: %w", err)
	}

	imageRefs, err := ExtractImageReferencesFromGitlabCi(pipeline)
	if err != nil {
		return fmt.Errorf("failed to extract image references: %w", err)
	}
	if len(imageRefs.Refs) == 0 {
		logging.Warning.Printfln("No image references found in GitLab CI file %s", filename)
		return nil
	}

	err = ReplaceInFile(filename, &imageRefs, tools)
	if err != nil {
		return fmt.Errorf("failed to bump image references: %w", err)
	}

	return nil
}
