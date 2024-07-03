package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/uniget-org/cli/pkg/logging"

	"github.com/charmbracelet/glamour"
)

func initReleaseNotesCmd() {
	rootCmd.AddCommand(releaseNotesCmd)
}

var releaseNotesCmd = &cobra.Command{
	Use:     "release-notes",
	Short:   "Show release notes for a tool",
	Long:    header + "\nShow release notes for a tool",
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

		tool, err := tools.GetByName(args[0])
		if err != nil {
			return fmt.Errorf("failed to get tool: %s", err)
		}

		if tool.Renovate.Datasource != "github-releases" {
			return fmt.Errorf("release notes are only available for tools with github-releases datasource. %s has %s", tool.Name, tool.Renovate.Datasource)
		}
		
		// if environment variable GITHUB_TOKEN is set, print it
		if len(os.Getenv("GITHUB_TOKEN")) == 0 {
			logging.Warning.Printfln("GITHUB_TOKEN is not set. You may experience failed requests due to rate limiting.")
		}

		url := fmt.Sprintf("https://api.github.com/repos/%s/releases/tags/%s", tool.Renovate.Package, tool.Version)

		client := &http.Client{}
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %s", err)
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", fmt.Sprintf("%s/%s", projectName, version))
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("failed fetch body of GitHub release for tool %s: %s", tool.Name, err)
		}
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read body of GitHub release for tool %s: %s", tool.Name, err)
		}

		var result map[string]interface{}
		json.Unmarshal([]byte(bodyBytes), &result)

		out, err := glamour.Render(result["body"].(string), "dark")
		if err != nil {
			return fmt.Errorf("failed to render release notes: %s", err)
		}
		fmt.Print(out)

		return nil
	},
}
