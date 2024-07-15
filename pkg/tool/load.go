package tool

import (
	"encoding/json"
	"fmt"
	"os"
)

func LoadFromFile(filename string) (Tools, error) {
	data, err := os.ReadFile(filename) // #nosec G304 -- filename is built when LoadFromFile is called
	if err != nil {
		return Tools{}, fmt.Errorf("error loading file contents: %s", err)
	}

	tools, err := LoadFromBytes(data)
	if err != nil {
		return Tools{}, fmt.Errorf("error loading data: %s", err)
	}

	return tools, nil
}

func LoadFromBytes(data []byte) (Tools, error) {
	var tools Tools

	err := json.Unmarshal(data, &tools)
	if err != nil {
		return Tools{}, err
	}

	for index, tool := range tools.Tools {
		if tool.Binary == "" {
			tools.Tools[index].Binary = fmt.Sprintf("${target}/bin/%s", tool.Name)
		}

		if tool.SchemaVersion == "" {
			tools.Tools[index].SchemaVersion = "1"
		}
	}

	return tools, nil
}
