package main

import (
	"fmt"
	"os"

	"github.com/nicholasdille/docker-setup/pkg/tool"
)

func main() {
	tools, err := tool.LoadFromFile("metadata.json")
	if err != nil {
		os.Exit(1)
	}

	//tools.List()
	tool, err := tools.GetByName("docker")
	if err != nil {
		os.Exit(1)
	}
	tool.ReplaceVariables("/usr/local", "x86_64", "amd64")

	err = tool.GetBinaryStatus()
	if err != nil {
		fmt.Printf("Failed to get binary status: %s", err)
		os.Exit(1)
	}

	err = tool.GetVersionStatus()
	if err != nil {
		fmt.Printf("Failed to get version status: %s", err)
		os.Exit(1)
	}

	tool.Print()
}
