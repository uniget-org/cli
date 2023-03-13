package main

import (
	"os"

	"github.com/nicholasdille/docker-setup/pkg/tool"
)

func main() {
	tools, err := tool.LoadFromFile("metadata.json")
	if err != nil {
		os.Exit(1)
	}

	tools.List()
	tools.Tools[0].Print()
}
