package main

import (
	"os"

	"github.com/nicholasdille/docker-setup/pkg/tool"
)

func main() {
	tools, err := tool.LoadFromFile2("metadata.json")
	if err != nil {
		os.Exit(1)
	}

	tools.List()
	tool, err := tools.GetByName("az")
	if err != nil {
		os.Exit(1)
	}
	tool.Print()
}
