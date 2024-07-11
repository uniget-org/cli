package tool

import (
	"fmt"
	"os"
	"strings"

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
	t.SetColumnConfigs([]table.ColumnConfig{
		{
			Number:   4,
			WidthMax: 80,
		},
	})
	t.Style().Options.DrawBorder = false
	t.Style().Options.SeparateColumns = false
	t.Style().Options.SeparateFooter = false
	t.Style().Options.SeparateHeader = false
	t.Style().Options.SeparateRows = false

	t.AppendHeader(table.Row{"Name", "Version", "Description"})

	for index, tool := range tools.Tools {
		t.AppendRows([]table.Row{
			{index + 1, tool.Name, tool.Version, tool.Description},
		})
	}

	t.Render()
}

func (tools *Tools) ListWithStatus() {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)

	t.AppendHeader(table.Row{"#", "Name", "Version", "Binary?", "Installed", "Matches?", "Skip?", "IsReq?"})

	for index, tool := range tools.Tools {
		t.AppendRows([]table.Row{
			{index + 1, tool.Name, tool.Version, tool.Status.BinaryPresent, tool.Status.Version, tool.Status.VersionMatches, tool.Status.SkipDueToConflicts || !tool.Status.IsRequested, tool.Status.IsRequested},
		})
	}

	t.Render()
}

func (tool *Tool) ShowInternals(indentation int) string {
	result := ""
	for _, line := range strings.Split(tool.Messages.Internals, "\n") {
		if line == "" {
			continue
		}
		result += fmt.Sprintf("%s%s\n", strings.Repeat(" ", indentation), line)
	}

	return result
}

func (tool *Tool) ShowUsage(indentation int) string {
	result := ""
	for _, line := range strings.Split(tool.Messages.Usage, "\n") {
		if line == "" {
			continue
		}
		result += fmt.Sprintf("%s%s\n", strings.Repeat(" ", indentation), line)
	}

	return result
}

func (tool *Tool) ShowUpdate(indentation int) string {
	result := ""
	for _, line := range strings.Split(tool.Messages.Update, "\n") {
		if line == "" {
			continue
		}
		result += fmt.Sprintf("%s%s\n", strings.Repeat(" ", indentation), line)
	}

	return result
}

func (tool *Tool) Print() {
	fmt.Printf("Name: %s\n", tool.Name)
	fmt.Printf("  Description: %s\n", tool.Description)
	if len(tool.Homepage) > 0 {
		fmt.Printf("  Homepage: %s\n", tool.Homepage)
	}
	fmt.Printf("  Repository: %s\n", tool.Repository)
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

	if tool.ConflictsWith != nil {
		fmt.Printf("  Conflicts with:\n")
		for _, conflict := range tool.ConflictsWith {
			fmt.Printf("    %s\n", conflict)
		}
	}

	if tool.Platforms != nil {
		fmt.Printf("  Platforms:\n")
		for _, dep := range tool.Platforms {
			fmt.Printf("    %s\n", dep)
		}
	}

	fmt.Printf("  Messages:\n")
	if tool.Messages.Internals != "" {
		fmt.Println("    Internals:")
		fmt.Print(tool.ShowInternals(6))
	}
	if tool.Messages.Usage != "" {
		fmt.Println("    Usage:")
		fmt.Print(tool.ShowUsage(6))
	}
	if tool.Messages.Update != "" {
		fmt.Println("    Update:")
		fmt.Print(tool.ShowUpdate(6))
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
	fmt.Printf("    Version: %s\n", tool.Status.Version)
	fmt.Printf("    Version matches: %t\n", tool.Status.VersionMatches)
	fmt.Printf("    Marker file present: %t\n", tool.Status.MarkerFilePresent)
	fmt.Printf("    Marker file version: %s\n", tool.Status.MarkerFileVersion)
	fmt.Printf("    Skip: %t\n", tool.Status.SkipDueToConflicts || !tool.Status.IsRequested)
	fmt.Printf("    Is requested: %t\n", tool.Status.IsRequested)
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
