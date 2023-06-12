package tool

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

func LoadFromFile(filename string) (Tools, error) {
	data, err := ioutil.ReadFile(filename)
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

	err := yaml.Unmarshal(data, &tools)
	if err != nil {
		return Tools{}, err
	}

	for index, tool := range tools.Tools {
		if tool.Binary == "" {
			tools.Tools[index].Binary = fmt.Sprintf("${target}/bin/%s", tool.Name)
		}
	}

	return tools, nil
}