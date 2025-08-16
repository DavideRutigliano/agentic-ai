package tools

import (
	"encoding/json"

	"github.com/invopop/jsonschema"
	"google.golang.org/genai"
)

func GenerateSchema(v any) *genai.Schema {
	s := jsonschema.Reflect(v)

	b, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}

	var genaiSchema genai.Schema
	err = json.Unmarshal(b, &genaiSchema)
	if err != nil {
		panic(err)
	}

	return &genaiSchema
}
