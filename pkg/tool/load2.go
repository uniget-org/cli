package tool

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

func LoadFromFile2(filename string) ([]Tool, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return []Tool{}, fmt.Errorf("Error loading file contents: %s\n", err)
	}

	tools, err := LoadFromBytes2(data)
	if err != nil {
		return []Tool{}, fmt.Errorf("Error loading data: %s\n", err)
	}

	return tools, nil
}

func LoadFromBytes2(data []byte) ([]Tool, error) {
	var tools Tools

	err := yaml.Unmarshal(data, &tools)
	if err != nil {
		return []Tool{}, err
	}

	for index, tool := range tools.Tools {
		if tool.Binary == "" {
			tools.Tools[index].Binary = fmt.Sprintf("${target}/bin/%s", tool.Name)
		}
	}

	return tools.Tools, nil
}