package tools

import (
	"encoding/json"
	"fmt"
	"os"

	"google.golang.org/genai"
)

type ReadFileInput struct {
	Path string `json:"path" jsonschema:"description=The path to the file to read."`
}

var ReadFileTool = &ToolDefinition{
	Declaration: &genai.FunctionDeclaration{
		Name:        "read_file",
		Description: "Reads the content of a file at the given path.",
		Parameters:  GenerateSchema(&ReadFileInput{}),
	},
	Function: func(args json.RawMessage) (string, error) {
		var input ReadFileInput
		if err := json.Unmarshal(args, &input); err != nil {
			return "", fmt.Errorf("failed to unmarshal args: %w", err)
		}
		if input.Path == "" {
			return "", fmt.Errorf("path is required")
		}

		data, err := os.ReadFile(input.Path)
		if err != nil {
			return "", fmt.Errorf("failed to read file: %w", err)
		}
		return string(data), nil
	},
}
