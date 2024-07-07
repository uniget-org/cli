package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sort"

	"gopkg.in/yaml.v3"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/uniget-org/cli/pkg/logging"
)

var describeOutput string
var versions bool
var upstreamVersions bool

func initDescribeCmd() {
	describeCmd.Flags().BoolVar(&versions, "versions", false, "Find available versions")
	describeCmd.Flags().BoolVar(&upstreamVersions, "upstream-versions", false, "Find upstream versions")
	describeCmd.Flags().StringVarP(&describeOutput, "output", "o", "pretty", "Output options: pretty, json, yaml")

	rootCmd.AddCommand(describeCmd)
}

var describeCmd = &cobra.Command{
	Use:     "describe",
	Aliases: []string{"d", "info"},
	Short:   "Show detailed information about tools",
	Long:    header + "\nShow detailed information about tools",
	Args:    cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return tools.GetNames(), cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if viper.GetBool("update") {
			err := downloadMetadata()
			if err != nil {
				return fmt.Errorf("error downloading metadata: %s", err)
			}
		}
		assertMetadataFileExists()
		assertMetadataIsLoaded()

		toolName := args[0]
		tool, err := tools.GetByName(toolName)
		if err != nil {
			return fmt.Errorf("error getting tool %s", toolName)
		}
		tool.ReplaceVariables(viper.GetString("prefix")+"/"+viper.GetString("target"), arch, altArch)
		err = tool.GetMarkerFileStatus(viper.GetString("prefix") + "/" + cacheDirectory)
		if err != nil {
			return fmt.Errorf("error getting marker file status: %s", err)
		}
		err = tool.GetBinaryStatus()
		if err != nil {
			return fmt.Errorf("error getting binary status: %s", err)
		}
		err = tool.GetVersionStatus()
		if err != nil {
			return fmt.Errorf("error getting version status: %s", err)
		}

		if describeOutput == "pretty" {
			tool.Print()

		} else if describeOutput == "json" {
			data, err := json.Marshal(tool)
			if err != nil {
				return fmt.Errorf("failed to marshal to json: %s", err)
			}
			fmt.Println(string(data))

		} else if describeOutput == "yaml" {
			yamlEncoder := yaml.NewEncoder(os.Stdout)
			yamlEncoder.SetIndent(2)
			defer yamlEncoder.Close()
			err := yamlEncoder.Encode(tool)
			if err != nil {
				return fmt.Errorf("failed to encode yaml: %s", err)
			}

		} else {
			return fmt.Errorf("invalid output format: %s", describeOutput)
		}

		if versions {
			// https://github.com/regclient/regclient/blob/main/cmd/regctl/tag.go#L101
			logging.Warning.Printfln("Available versions not implemented yet.")
		}

		if upstreamVersions {
			var releaseTags []string
			switch tool.Renovate.Datasource {
				case "github-releases":
					releaseTags, err = fetchGitHubReleases(tool.Renovate.Package)
					if err != nil {
						return fmt.Errorf("failed to fetch GitHub releases: %s", err)
					}

				case "gitlab-releases":
					releaseTags, err = fetchGitLabReleases(tool.Renovate.Package)
					if err != nil {
						return fmt.Errorf("failed to fetch GitLab releases: %s", err)
					}
	
				case "gitea-releases":
					releaseTags, err = fetchGiteaReleases(tool.Renovate.Package)
					if err != nil {
						return fmt.Errorf("failed to fetch Gitea releases: %s", err)
					}

				case "npm":
					releaseTags, err = fetchNpmReleases(tool.Renovate.Package)
					if err != nil {
						return fmt.Errorf("failed to fetch Gitea releases: %s", err)
					}

				case "pypi":
					releaseTags, err = fetchPypiReleases(tool.Renovate.Package)
					if err != nil {
						return fmt.Errorf("failed to fetch Gitea releases: %s", err)
					}

				default:
					logging.Warning.Printfln("Upstream versions are not available for datasource %s", tool.Renovate.Datasource)
			}

			if len(releaseTags) > 0 {
				fmt.Println("  Upstream versions (most recent):")
				sort.Sort(sort.Reverse(byVersion(releaseTags)))
				for _, releaseTag := range releaseTags {
					version, err := extractVersionfromTag(releaseTag, tool.Renovate.ExtractVersion)
					if err != nil {
						return fmt.Errorf("failed to extract version from tag: %s", err)
					}
					fmt.Printf("    %s\n", version)
				}
	
			} else {
				logging.Warning.Printfln("  No upstream versions found")
			}
		}

		return nil
	},
}

func extractVersionfromTag(tag string, regex string) (string, error) {
	if len(regex) > 0 {
		re, err := regexp.Compile(regex)
		if err != nil {
			return "", fmt.Errorf("cannot compile regexp: %w", err)
		}
		matches := re.FindStringSubmatch(tag)
		if len(matches) > 1 {
			return matches[re.SubexpIndex("version")], nil
		}
	}

	return tag, nil
}

func fetchGitHubReleases(project string) ([]string, error) {
	if len(os.Getenv("GITHUB_TOKEN")) == 0 {
		logging.Warning.Printfln("GITHUB_TOKEN is not set. You may experience failed requests due to rate limiting.")
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/releases", project)
	logging.Debugf("Fetching releases from %s", url)

	bodyBytes, err := fetchUrl(url)
	if err != nil {
		return []string{}, fmt.Errorf("failed to fetch body of GitHub release: %s", err)
	}

	var releases []interface{}
	err = json.Unmarshal(bodyBytes, &releases)
	if err != nil {
		return []string{}, fmt.Errorf("failed to parse body of GitHub releases: %s", err)
	}

	var releaseTags []string = make([]string, 0)
	for index := range releases {
		release := releases[index].(map[string]interface{})
		releaseTags = append(releaseTags, release["tag_name"].(string))
	}

	return releaseTags, nil
}

func fetchGitLabReleases(project string) ([]string, error) {
	projectUrlEncoded := strings.ReplaceAll(project, "/", "%2f")
	url := fmt.Sprintf("https://gitlab.com/api/v4/projects/%s/releases", projectUrlEncoded)
	logging.Debugf("Fetching releases from %s", url)

	bodyBytes, err := fetchUrl(url)
	if err != nil {
		return []string{}, fmt.Errorf("failed to fetch body of GitLab release: %s", err)
	}

	var releases []interface{}
	err = json.Unmarshal(bodyBytes, &releases)
	if err != nil {
		return []string{}, fmt.Errorf("failed to parse body of GitLab releases: %s", err)
	}

	var releaseTags []string = make([]string, 0)
	for index := range releases {
		release := releases[index].(map[string]interface{})
		releaseTags = append(releaseTags, release["tag_name"].(string))
	}

	return releaseTags, nil
}

func fetchGiteaReleases(project string) ([]string, error) {
	url := fmt.Sprintf("https://gitea.com/api/v1/repos/%s/releases", project)
	logging.Debugf("Fetching releases from %s", url)

	bodyBytes, err := fetchUrl(url)
	if err != nil {
		return []string{}, fmt.Errorf("failed to fetch body of Gitea release: %s", err)
	}

	var releases []interface{}
	err = json.Unmarshal(bodyBytes, &releases)
	if err != nil {
		return []string{}, fmt.Errorf("failed to parse body of Gitea releases: %s", err)
	}

	var releaseTags []string = make([]string, 0)
	for index := range releases {
		release := releases[index].(map[string]interface{})
		releaseTags = append(releaseTags, release["tag_name"].(string))
	}

	return releaseTags, nil
}

func fetchNpmReleases(project string) ([]string, error) {
	url := fmt.Sprintf("https://registry.npmjs.com/%s", project)
	logging.Debugf("Fetching releases from %s", url)

	bodyBytes, err := fetchUrl(url)
	if err != nil {
		return []string{}, fmt.Errorf("failed to fetch body of npm release: %s", err)
	}

	var npmPackage map[string]interface{}
	err = json.Unmarshal(bodyBytes, &npmPackage)
	if err != nil {
		return []string{}, fmt.Errorf("failed to parse body of npm releases: %s", err)
	}

	var versionTags []string = make([]string, 0)
	versions := npmPackage["versions"].(map[string]interface{})
	for versionTag := range versions {
		versionTags = append(versionTags, versionTag)
	}

	return versionTags, nil
}

func fetchPypiReleases(project string) ([]string, error) {
	url := fmt.Sprintf("https://pypi.org/pypi/%s/json", project)
	logging.Debugf("Fetching releases from %s", url)

	bodyBytes, err := fetchUrl(url)
	if err != nil {
		return []string{}, fmt.Errorf("failed to fetch body of pypi release: %s", err)
	}

	var pypiPackage map[string]interface{}
	err = json.Unmarshal(bodyBytes, &pypiPackage)
	if err != nil {
		return []string{}, fmt.Errorf("failed to parse body of pypi releases: %s", err)
	}

	var versionTags []string = make([]string, 0)
	versions := pypiPackage["releases"].(map[string]interface{})
	for versionTag := range versions {
		versionTags = append(versionTags, versionTag)
	}

	return versionTags, nil
}
