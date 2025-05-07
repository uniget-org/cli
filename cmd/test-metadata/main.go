package main

import (
	"fmt"
	"os"

	"github.com/uniget-org/cli/pkg/tool"
)

func main() {
	metadataFile := "./metadata.json"

	tools, err := tool.LoadFromFile(metadataFile)
	if err != nil {
		panic(fmt.Errorf("failed to load metadata from file %s: %s", metadataFile, err))
	}

	tools.List(os.Stdout)
}
