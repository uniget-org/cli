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
	//nolint:errcheck
	fmt.Fprintf(w, "Name: %s\n", tool.Name)
	if len(tool.SchemaVersion) > 0 {
		//nolint:errcheck
		fmt.Fprintf(w, "  Schema version: %s\n", tool.SchemaVersion)
	}
	//nolint:errcheck
	fmt.Fprintf(w, "  Description: %s\n", tool.Description)
	//nolint:errcheck
	fmt.Fprintf(w, "  Homepage: %s\n", tool.Homepage)
	//nolint:errcheck
	fmt.Fprintf(w, "  Repository: %s\n", tool.Repository)
	//nolint:errcheck
	fmt.Fprintf(w, "  License: %s", tool.License.Name)
	if len(tool.License.Link) > 0 {
		//nolint:errcheck
		fmt.Fprintf(w, " (%s)", tool.License.Link)
	}
	fmt.Print("\n")
	//nolint:errcheck
	fmt.Fprintf(w, "  Version: %s\n", tool.Version)

	if tool.Binary != "" {
		//nolint:errcheck
		fmt.Fprintf(w, "  Binary: %s\n", tool.Binary)
	}

	if len(tool.Check) > 0 {
		//nolint:errcheck
		fmt.Fprintf(w, "  Check: <%s>\n", tool.Check)
	}

	//nolint:errcheck
	fmt.Fprintf(w, "  Tags:\n")
	for _, tag := range tool.Tags {
		//nolint:errcheck
		fmt.Fprintf(w, "    %s\n", tag)
	}

	if tool.BuildDependencies != nil {
		//nolint:errcheck
		fmt.Fprintf(w, "  Build dependencies:\n")
		for _, dep := range tool.BuildDependencies {
			//nolint:errcheck
			fmt.Fprintf(w, "    %s\n", dep)
		}
	}

	if tool.RuntimeDependencies != nil {
		//nolint:errcheck
		fmt.Fprintf(w, "  Runtime dependencies:\n")
		for _, dep := range tool.RuntimeDependencies {
			//nolint:errcheck
			fmt.Fprintf(w, "    %s\n", dep)
		}
	}

	if tool.ConflictsWith != nil {
		//nolint:errcheck
		fmt.Fprintf(w, "  Conflicts with:\n")
		for _, conflict := range tool.ConflictsWith {
			//nolint:errcheck
			fmt.Fprintf(w, "    %s\n", conflict)
		}
	}

	if tool.Platforms != nil {
		//nolint:errcheck
		fmt.Fprintf(w, "  Platforms:\n")
		for _, dep := range tool.Platforms {
			//nolint:errcheck
			fmt.Fprintf(w, "    %s\n", dep)
		}
	}

	//nolint:errcheck
	fmt.Fprintf(w, "  Messages:\n")
	if tool.Messages.Internals != "" {
		//nolint:errcheck
		fmt.Fprint(w, "    Internals:\n")
		//nolint:errcheck
		fmt.Fprint(w, tool.ShowInternals(6))
	}
	if tool.Messages.Usage != "" {
		//nolint:errcheck
		fmt.Fprint(w, "    Usage:")
		//nolint:errcheck
		fmt.Fprint(w, tool.ShowUsage(6))
	}
	if tool.Messages.Update != "" {
		//nolint:errcheck
		fmt.Fprint(w, "    Update:")
		//nolint:errcheck
		fmt.Fprint(w, tool.ShowUpdate(6))
	}

	if tool.Renovate.Datasource != "" {
		//nolint:errcheck
		fmt.Fprintf(w, "  Renovate:\n")
		//nolint:errcheck
		fmt.Fprintf(w, "    Datasource: %s\n", tool.Renovate.Datasource)
		//nolint:errcheck
		fmt.Fprintf(w, "    Package: %s\n", tool.Renovate.Package)
		if tool.Renovate.ExtractVersion != "" {
			//nolint:errcheck
			fmt.Fprintf(w, "    ExtractVersion: %s\n", tool.Renovate.ExtractVersion)
		}
		if tool.Renovate.Versioning != "" {
			//nolint:errcheck
			fmt.Fprintf(w, "    Versioning: %s\n", tool.Renovate.Versioning)
		}
	}

	if len(tool.Sources) > 0 {
		//nolint:errcheck
		fmt.Fprintf(w, "  Sources:\n")
		for _, source := range tool.Sources {
			//nolint:errcheck
			fmt.Fprintf(w, "    %s/%s\n", source.Registry, source.Repository)
		}
	}

	//nolint:errcheck
	fmt.Fprintf(w, "  Status\n")
	//nolint:errcheck
	fmt.Fprintf(w, "    Binary present: %t\n", tool.Status.BinaryPresent)
	//nolint:errcheck
	fmt.Fprintf(w, "    Version: %s\n", tool.Status.Version)
	//nolint:errcheck
	fmt.Fprintf(w, "    Version matches: %t\n", tool.Status.VersionMatches)
	//nolint:errcheck
	fmt.Fprintf(w, "    Marker file present: %t\n", tool.Status.MarkerFilePresent)
	//nolint:errcheck
	fmt.Fprintf(w, "    Marker file version: %s\n", tool.Status.MarkerFileVersion)
	//nolint:errcheck
	fmt.Fprintf(w, "    Skip: %t\n", tool.Status.SkipDueToConflicts || !tool.Status.IsRequested)
	//nolint:errcheck
	fmt.Fprintf(w, "    Is requested: %t\n", tool.Status.IsRequested)
}

func (tools *Tools) Describe(w io.Writer, name string) error {
	for _, tool := range tools.Tools {
		if tool.Name == name {
			_, err := fmt.Fprintf(w, "%+v\n", tool)
			if err != nil {
				return fmt.Errorf("error printing tool information: %s", err)
			}
			return nil
		}
	}

	return fmt.Errorf("tool named %s not found", name)
}
