package tool

import (
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
)

func (tool *Tool) List() {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)

	t.AppendHeader(table.Row{"#", "Name", "Version"})

	t.AppendRows([]table.Row{
		{1, tool.Name, tool.Version},
	})

	t.Render()
}

func (tools *Tools) List() {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.Style().Options.DrawBorder = false
	t.Style().Options.SeparateColumns = false
	t.Style().Options.SeparateFooter = false
	t.Style().Options.SeparateHeader = false
	t.Style().Options.SeparateRows = false

	t.AppendHeader(table.Row{"Name", "Version"})

	for index, tool := range tools.Tools {
		t.AppendRows([]table.Row{
			{index + 1, tool.Name, tool.Version},
		})
	}

	t.Render()
}

func (tools *Tools) ListWithStatus() {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)

	t.AppendHeader(table.Row{"#", "Name", "Version", "Binary?", "Installed", "Matches?"})

	for index, tool := range tools.Tools {
		t.AppendRows([]table.Row{
			{index + 1, tool.Name, tool.Version, tool.Status.BinaryPresent, tool.Status.Version, tool.Status.VersionMatches},
		})
	}

	t.Render()
}

func (tool *Tool) Print() {
	fmt.Printf("Name: %s\n", tool.Name)
	fmt.Printf("  %+v", tool)
	fmt.Printf("  Description: %s\n", tool.Description)
	fmt.Printf("  Homepage: %s\n", tool.Homepage)
	fmt.Printf("  Version: %s\n", tool.Version)

	if tool.Binary != "" {
		fmt.Printf("  Binary: %s\n", tool.Binary)
	}

	if len(tool.Check) > 0 {
		fmt.Printf("  Check: <%s>\n", tool.Check)
	}

	fmt.Printf("  Tags:\n")
	for _, tag := range tool.Tags {
		fmt.Printf("    %s\n", tag)
	}

	if tool.BuildDependencies != nil {
		fmt.Printf("  Build dependencies:\n")
		for _, dep := range tool.BuildDependencies {
			fmt.Printf("    %s\n", dep)
		}
	}

	if tool.RuntimeDependencies != nil {
		fmt.Printf("  Runtime dependencies:\n")
		for _, dep := range tool.RuntimeDependencies {
			fmt.Printf("    %s\n", dep)
		}
	}

	if tool.Platforms != nil {
		fmt.Printf("  Platforms:\n")
		for _, dep := range tool.Platforms {
			fmt.Printf("    %s\n", dep)
		}
	}

	if tool.Renovate.Datasource != "" {
		fmt.Printf("  Renovate:\n")
		fmt.Printf("    Datasource: %s\n", tool.Renovate.Datasource)
		fmt.Printf("    Package: %s\n", tool.Renovate.Package)
		if tool.Renovate.ExtractVersion != "" {
			fmt.Printf("    ExtractVersion: %s\n", tool.Renovate.ExtractVersion)
		}
		if tool.Renovate.Versioning != "" {
			fmt.Printf("    Versioning: %s\n", tool.Renovate.Versioning)
		}
	}

	fmt.Printf("  Status\n")
	fmt.Printf("    Binary present: %t\n", tool.Status.BinaryPresent)
	if tool.Status.Version != "" {
		fmt.Printf("    Version: %s\n", tool.Status.Version)
		fmt.Printf("    Version matches: %t\n", tool.Status.VersionMatches)
	}
}

func (tools *Tools) Describe(name string) error {
	for _, tool := range tools.Tools {
		if tool.Name == name {
			fmt.Printf("%+v\n", tool)
			return nil
		}
	}

	return fmt.Errorf("Tool named %s not found", name)
}
