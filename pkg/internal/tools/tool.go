package tools

import (
	"encoding/json"

	"google.golang.org/genai"
)

type ToolDefinition struct {
	Declaration *genai.FunctionDeclaration
	Function func(args json.RawMessage) (string, error)
}
