package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/uniget-org/cli/pkg/logging"

	"github.com/charmbracelet/glamour"
)

var releaseNotesVersion string

func initReleaseNotesCmd() {
	releaseNotesCmd.Flags().StringVar(&releaseNotesVersion, "version", "", "Show release notes for a specific version")

	rootCmd.AddCommand(releaseNotesCmd)
}

var releaseNotesCmd = &cobra.Command{
	Use:   "release-notes",
	Short: "Show release notes for a tool",
	Long:  header + "\nShow release notes for a tool",
	Args:  cobra.ExactArgs(1),
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

		tool, err := tools.GetByName(args[0])
		if err != nil {
			return fmt.Errorf("failed to get tool: %s", err)
		}
		checkClientVersionRequirement(tool)

		var payload []byte
		var bodyFieldName string
		versionTag := tool.Version
		if len(releaseNotesVersion) > 0 {
			logging.Debugf("Using version %s for release notes", releaseNotesVersion)
			versionTag = releaseNotesVersion
		}
		if tool.Renovate.ExtractVersion != "" {
			re, err := regexp.Compile(`\(\?[^)]+\)`)
			if err != nil {
				return fmt.Errorf("cannot compile regexp: %w", err)
			}
			versionTag = re.ReplaceAllString(tool.Renovate.ExtractVersion, versionTag)
			versionTag = strings.Replace(versionTag, "^", "", -1)
			versionTag = strings.Replace(versionTag, "$", "", -1)
		}
		logging.Debugf("Using version tag %s for release notes", versionTag)
		switch tool.Renovate.Datasource {
		case "github-releases":
			payload, err = fetchBodyFromGitHubRelease(tool.Renovate.Package, versionTag)
			if err != nil {
				return fmt.Errorf("failed to fetch body of GitHub release for tool %s: %s", tool.Name, err)
			}
			bodyFieldName = "body"

		case "gitlab-releases":
			payload, err = fetchBodyFromGitLabRelease(tool.Renovate.Package, versionTag)
			if err != nil {
				return fmt.Errorf("failed to fetch body of GitLab release for tool %s: %s", tool.Name, err)
			}
			bodyFieldName = "description"

		case "gitea-releases":
			payload, err = fetchBodyFromGiteaRelease(tool.Renovate.Package, versionTag)
			if err != nil {
				return fmt.Errorf("failed to fetch body of Gitea release for tool %s: %s", tool.Name, err)
			}
			bodyFieldName = "body"

		case "npm":
			payload, err = fetchBodyFromNpm(tool.Renovate.Package, versionTag)
			if err != nil {
				return fmt.Errorf("failed to fetch body from npm for tool %s: %s", tool.Name, err)
			}
			bodyFieldName = "body"

		case "pypi":
			payload, err = fetchBodyFromPypi(tool.Renovate.Package, versionTag)
			if err != nil {
				return fmt.Errorf("failed to fetch body from pypi for tool %s: %s", tool.Name, err)
			}
			bodyFieldName = "body"

		default:
			return fmt.Errorf("release notes are not available for datasource %s", tool.Renovate.Datasource)
		}

		var result map[string]interface{}
		err = json.Unmarshal(payload, &result)
		if err != nil {
			return fmt.Errorf("failed to parse body of GitHub release for tool %s: %s", tool.Name, err)
		}

		out, err := glamour.Render(result[bodyFieldName].(string), "dark")
		if err != nil {
			return fmt.Errorf("failed to render release notes: %s", err)
		}
		fmt.Print(out)

		return nil
	},
}

func fetchUrl(url string) ([]byte, error) {
	logging.Debugf("Fetching %s", url)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to create request: %s", err)
	}
	req.Header.Set("User-Agent", fmt.Sprintf("%s/%s", projectName, version))
	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, fmt.Errorf("failed fetch url: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return []byte{}, fmt.Errorf("failed to fetch url: %s", resp.Status)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to read url: %s", err)
	}

	return bodyBytes, nil
}

func fetchBodyFromGitHubRelease(project string, versionTag string) ([]byte, error) {
	if len(os.Getenv("GITHUB_TOKEN")) == 0 {
		logging.Warning.Printfln("GITHUB_TOKEN is not set. You may experience failed requests due to rate limiting.")
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/tags/%s", project, versionTag)
	logging.Debugf("Fetching release notes from %s", url)

	bodyBytes, err := fetchUrl(url)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to fetch body of GitHub release: %s", err)
	}

	return bodyBytes, nil
}

func fetchBodyFromGitLabRelease(project string, versionTag string) ([]byte, error) {
	projectUrlEncoded := strings.ReplaceAll(project, "/", "%2f")
	url := fmt.Sprintf("https://gitlab.com/api/v4/projects/%s/releases/%s", projectUrlEncoded, versionTag)
	logging.Debugf("Fetching release notes from %s", url)

	bodyBytes, err := fetchUrl(url)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to fetch body of GitLab release: %s", err)
	}

	return bodyBytes, nil
}

func fetchBodyFromGiteaRelease(project string, versionTag string) ([]byte, error) {
	url := fmt.Sprintf("https://gitea.com/api/v1/repos/%s/releases/tags/%s", project, versionTag)
	logging.Debugf("Fetching release notes from %s", url)

	bodyBytes, err := fetchUrl(url)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to fetch body of Gitea release: %s", err)
	}

	return bodyBytes, nil
}

func extractGitHubOwnerAndProject(url string) (string, error) {
	re, err := regexp.Compile(`^(git\+)?https://github\.com/(?<owner>[^/]+)/(?<project>[^/.]+)(\.git)?$`)
	if err != nil {
		return "", fmt.Errorf("cannot compile regexp: %w", err)
	}
	matches := re.FindStringSubmatch(url)
	owner := matches[re.SubexpIndex("owner")]
	project := matches[re.SubexpIndex("project")]

	return fmt.Sprintf("%s/%s", owner, project), nil
}

func fetchBodyFromNpm(project string, versionTag string) ([]byte, error) {
	url := fmt.Sprintf("https://registry.npmjs.com/%s/%s", project, versionTag)
	logging.Debugf("Fetching release notes from %s", url)

	bodyBytes, err := fetchUrl(url)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to fetch body from npm: %s", err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(bodyBytes, &result)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to parse npm response: %s", err)
	}

	repo := result["repository"].(map[string]interface{})
	if repo["type"] == "git" && strings.Contains(repo["url"].(string), "github.com") {
		project, err := extractGitHubOwnerAndProject(repo["url"].(string))
		if err != nil {
			return []byte{}, fmt.Errorf("failed to fetch body of GitHub release for npm package: %s", err)
		}

		GhBody, err := fetchBodyFromGitHubRelease(project, versionTag)
		if err != nil {
			return []byte{}, fmt.Errorf("failed to fetch body of GitHub release for npm package: %s", err)
		}

		return GhBody, nil

	} else {
		return []byte{}, fmt.Errorf("unsupported repository type <%s> for npm package %s", repo["type"], project)
	}
}

func fetchBodyFromPypi(project string, versionTag string) ([]byte, error) {
	url := fmt.Sprintf("https://pypi.org/pypi/%s/%s/json", project, versionTag)
	logging.Debugf("Fetching release notes from %s", url)

	bodyBytes, err := fetchUrl(url)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to fetch body from pypi: %s", err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(bodyBytes, &result)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to parse pypi response: %s", err)
	}

	info := result["info"].(map[string]interface{})
	urls := info["project_urls"].(map[string]interface{})
	if urls["Homepage"] != nil {
		project, err := extractGitHubOwnerAndProject(urls["Homepage"].(string))
		if err != nil {
			return []byte{}, fmt.Errorf("failed to extract owner/project from homepage <%s>: %s", urls["Homepage"].(string), err)
		}

		GhBody, err := fetchBodyFromGitHubRelease(project, versionTag)
		if err != nil {
			return []byte{}, fmt.Errorf("failed to fetch body of GitHub release for npm package: %s", err)
		}

		return GhBody, nil

	} else {
		return []byte{}, fmt.Errorf("no homepage found for pypi package %s", project)
	}
}
