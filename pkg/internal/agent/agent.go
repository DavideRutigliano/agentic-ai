package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"google.golang.org/genai"

	"daviderutigliano/agentic-ai/internal/tools"
)

func NewAgent(client *genai.Client, getUserMessage func() (string, bool), activeTools []*tools.ToolDefinition, verbose bool) *Agent {
	toolMap := make(map[string]*tools.ToolDefinition)
	for _, t := range activeTools {
		toolMap[t.Declaration.Name] = t
	}

	return &Agent{
		client:         client,
		getUserMessage: getUserMessage,
		tools:          toolMap,
		verbose:        verbose,
	}
}

type Agent struct {
	client         *genai.Client
	getUserMessage func() (string, bool)
	tools          map[string]*tools.ToolDefinition
	verbose        bool
}

func (a *Agent) Run(ctx context.Context) error {
	conversation := []*genai.Content{}
	var functionDeclarations []*genai.FunctionDeclaration
	for _, t := range a.tools {
		functionDeclarations = append(functionDeclarations, t.Declaration)
	}
	var availableTools []*genai.Tool
	if len(functionDeclarations) > 0 {
		availableTools = []*genai.Tool{{FunctionDeclarations: functionDeclarations}}
	}

	if a.verbose {
		log.Println("Starting chat session")
	}
	fmt.Println("Chat with Gemini (use 'ctrl-c' to quit)")

	for {
		fmt.Print("\u001b[94mYou\u001b[0m: ")
		userInput, ok := a.getUserMessage()
		if !ok {
			if a.verbose {
				log.Println("User input ended, breaking from chat loop")
			}
			break
		}

		if userInput == "" {
			if a.verbose {
				log.Println("Skipping empty message")
			}
			continue
		}

		if a.verbose {
			log.Printf("User input received: %q", userInput)
		}

		userMessage := []*genai.Part{genai.NewPartFromText(userInput)}
		conversation = append(conversation, genai.NewContentFromParts(userMessage, genai.RoleUser))

		for {
			if a.verbose {
				log.Printf("Sending message to Gemini, conversation length: %d", len(conversation))
			}

			resp, err := a.runInference(ctx, conversation, availableTools)
			if err != nil {
				if a.verbose {
					log.Printf("Error during inference: %v", err)
				}
				return err
			}

			if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
				fmt.Println("\u001b[93mGemini\u001b[0m: I am unable to answer.")
				if resp.PromptFeedback != nil {
					if a.verbose {
						log.Printf("Prompt feedback: %v", resp.PromptFeedback)
					}
					fmt.Printf("Blocked for reason: %s\n", resp.PromptFeedback.BlockReason)
				}
				break
			}

			conversation = append(conversation, resp.Candidates[0].Content)

			if text := resp.Text(); text != "" {
				fmt.Printf("\u001b[93mGemini\u001b[0m: %s\n", text)
			}

			funcCalls := resp.FunctionCalls()
			if len(funcCalls) == 0 {
				break
			}

			if a.verbose {
				log.Printf("Received %d tool calls.", len(funcCalls))
			}

			toolResponseContent := a.executeTools(funcCalls)
			conversation = append(conversation, toolResponseContent)

			if a.verbose {
				log.Println("Sending tool responses back to Gemini.")
			}
		}
	}

	if a.verbose {
		log.Println("Chat session ended")
	}
	return nil
}

func (a *Agent) executeTools(calls []*genai.FunctionCall) *genai.Content {
	var toolParts []*genai.Part
	for _, call := range calls {
		if a.verbose {
			log.Printf("Executing tool: %s with args: %v", call.Name, call.Args)
		}

		tool, ok := a.tools[call.Name]
		if !ok {
			log.Printf("Error: unknown tool %q called by model", call.Name)
			toolParts = append(toolParts, genai.NewPartFromFunctionResponse(call.Name, map[string]any{
				"error": fmt.Sprintf("tool %q not found", call.Name),
			}))
			continue
		}

		argBytes, err := json.Marshal(call.Args)
		if err != nil {
			log.Printf("Error marshalling args for tool %q: %v", call.Name, err)
			toolParts = append(toolParts, genai.NewPartFromFunctionResponse(call.Name, map[string]any{
				"error": fmt.Sprintf("failed to marshal args: %v", err),
			}))
			continue
		}

		result, err := tool.Function(argBytes)
		if err != nil {
			log.Printf("Error executing tool %q: %v", call.Name, err)
			toolParts = append(toolParts, genai.NewPartFromFunctionResponse(call.Name, map[string]any{
				"error": fmt.Sprintf("tool execution failed: %v", err),
			}))
			continue
		}

		if a.verbose {
			log.Printf("Tool %q result: %s", call.Name, result)
		}
		toolParts = append(toolParts, genai.NewPartFromFunctionResponse(call.Name, map[string]any{
			"output": result,
		}))
	}
	return genai.NewContentFromParts(toolParts, "tool")
}

func (a *Agent) runInference(ctx context.Context, conversation []*genai.Content, tools []*genai.Tool) (*genai.GenerateContentResponse, error) {
	if a.verbose {
		log.Printf("Making API call to Gemini with model: %s", "gemini-2.5-flash")
	}

	var config *genai.GenerateContentConfig
	if len(tools) > 0 {
		config = &genai.GenerateContentConfig{
			Tools: tools,
		}
	}

	message, err := a.client.Models.GenerateContent(
		ctx,
		"gemini-2.5-flash",
		conversation,
		config,
	)

	if a.verbose {
		if err != nil {
			log.Printf("API call failed: %v", err)
		} else {
			log.Printf("API call successful, response received")
		}
	}

	return message, err
}
