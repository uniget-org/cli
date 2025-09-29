package parse

import (
	"fmt"
	"os"

	myos "github.com/uniget-org/cli/pkg/os"
	"github.com/uniget-org/cli/pkg/tool"
)

func ReplaceInFile(filename string, imageRefs *ImageRefs, tools *tool.Tools) error {
	file, err := myos.SlurpFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	err = imageRefs.Bump(tools)
	if err != nil {
		return fmt.Errorf("failed to bump image references: %w", err)
	}
	if len(imageRefs.Refs) != len(imageRefs.BumpedRefs) {
		return fmt.Errorf("mismatched refs (%d) and bumped refs (%d)", len(imageRefs.Refs), len(imageRefs.BumpedRefs))
	}

	file, err = imageRefs.Replace(file)
	if err != nil {
		return fmt.Errorf("failed to bump image references: %w", err)
	}

	stat, err := os.Stat(filename)
	if err != nil {
		return fmt.Errorf("failed to stat file %s: %w", filename, err)
	}
	err = os.WriteFile(filename, file, stat.Mode())
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", filename, err)
	}

	return nil
}
