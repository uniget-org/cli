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

	//tools.List()
	tool, err := tools.GetByName("regclient")
	if err != nil {
		os.Exit(1)
	}
	tool.GetBinaryStatus()
	tool.GetVersionStatus()
	tool.Print()
}
