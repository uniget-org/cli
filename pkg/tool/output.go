package tool

import (
	"fmt"
	"io"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
)

func (tool *Tool) List(w io.Writer) {
	t := table.NewWriter()
	t.SetOutputMirror(w)

	t.AppendHeader(table.Row{"#", "Name", "Version"})

	t.AppendRows([]table.Row{
		{1, tool.Name, tool.Version},
	})

	t.Render()
}

func (tools *Tools) List(w io.Writer) {
	t := table.NewWriter()
	t.SetOutputMirror(w)
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

	t.AppendHeader(table.Row{"#", "Name", "Version", "Description"})

	for index, tool := range tools.Tools {
		t.AppendRows([]table.Row{
			{index + 1, tool.Name, tool.Version, tool.Description},
		})
	}

	t.Render()
}

func (tools *Tools) ListWithStatus(w io.Writer) {
	t := table.NewWriter()
	t.SetOutputMirror(w)

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

func (tool *Tool) Print(w io.Writer) {
	fmt.Fprintf(w, "Name: %s\n", tool.Name)
	if len(tool.SchemaVersion) > 0 {
		fmt.Fprintf(w, "  Schema version: %s\n", tool.SchemaVersion)
	}
	fmt.Fprintf(w, "  Description: %s\n", tool.Description)
	fmt.Fprintf(w, "  Homepage: %s\n", tool.Homepage)
	fmt.Fprintf(w, "  Repository: %s\n", tool.Repository)
	fmt.Fprintf(w, "  License: %s", tool.License.Name)
	if len(tool.License.Link) > 0 {
		fmt.Fprintf(w, " (%s)", tool.License.Link)
	}
	fmt.Print("\n")
	fmt.Fprintf(w, "  Version: %s\n", tool.Version)

	if tool.Binary != "" {
		fmt.Fprintf(w, "  Binary: %s\n", tool.Binary)
	}

	if len(tool.Check) > 0 {
		fmt.Fprintf(w, "  Check: <%s>\n", tool.Check)
	}

	fmt.Fprintf(w, "  Tags:\n")
	for _, tag := range tool.Tags {
		fmt.Fprintf(w, "    %s\n", tag)
	}

	if tool.BuildDependencies != nil {
		fmt.Fprintf(w, "  Build dependencies:\n")
		for _, dep := range tool.BuildDependencies {
			fmt.Fprintf(w, "    %s\n", dep)
		}
	}

	if tool.RuntimeDependencies != nil {
		fmt.Fprintf(w, "  Runtime dependencies:\n")
		for _, dep := range tool.RuntimeDependencies {
			fmt.Fprintf(w, "    %s\n", dep)
		}
	}

	if tool.ConflictsWith != nil {
		fmt.Fprintf(w, "  Conflicts with:\n")
		for _, conflict := range tool.ConflictsWith {
			fmt.Fprintf(w, "    %s\n", conflict)
		}
	}

	if tool.Platforms != nil {
		fmt.Fprintf(w, "  Platforms:\n")
		for _, dep := range tool.Platforms {
			fmt.Fprintf(w, "    %s\n", dep)
		}
	}

	fmt.Fprintf(w, "  Messages:\n")
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
		fmt.Fprintf(w, "  Renovate:\n")
		fmt.Fprintf(w, "    Datasource: %s\n", tool.Renovate.Datasource)
		fmt.Fprintf(w, "    Package: %s\n", tool.Renovate.Package)
		if tool.Renovate.ExtractVersion != "" {
			fmt.Fprintf(w, "    ExtractVersion: %s\n", tool.Renovate.ExtractVersion)
		}
		if tool.Renovate.Versioning != "" {
			fmt.Fprintf(w, "    Versioning: %s\n", tool.Renovate.Versioning)
		}
	}

	fmt.Fprintf(w, "  Status\n")
	fmt.Fprintf(w, "    Binary present: %t\n", tool.Status.BinaryPresent)
	fmt.Fprintf(w, "    Version: %s\n", tool.Status.Version)
	fmt.Fprintf(w, "    Version matches: %t\n", tool.Status.VersionMatches)
	fmt.Fprintf(w, "    Marker file present: %t\n", tool.Status.MarkerFilePresent)
	fmt.Fprintf(w, "    Marker file version: %s\n", tool.Status.MarkerFileVersion)
	fmt.Fprintf(w, "    Skip: %t\n", tool.Status.SkipDueToConflicts || !tool.Status.IsRequested)
	fmt.Fprintf(w, "    Is requested: %t\n", tool.Status.IsRequested)
}

func (tools *Tools) Describe(w io.Writer, name string) error {
	for _, tool := range tools.Tools {
		if tool.Name == name {
			fmt.Fprintf(w, "%+v\n", tool)
			return nil
		}
	}

	return fmt.Errorf("Tool named %s not found", name)
}
