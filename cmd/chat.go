package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"google.golang.org/genai"
)

func main() {
	verbose := flag.Bool("verbose", false, "enable verbose logging")
	flag.Parse()

	if *verbose {
		log.SetOutput(os.Stderr)
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Println("Verbose logging enabled")
	} else {
		log.SetOutput(os.Stdout)
		log.SetFlags(0)
		log.SetPrefix("")
	}

	ctx := context.Background()
    // The client gets the API key from the environment variable `GEMINI_API_KEY`.
    client, err := genai.NewClient(ctx, nil)
    if err != nil {
        log.Fatal(err)
    }

	scanner := bufio.NewScanner(os.Stdin)
	getUserMessage := func() (string, bool) {
		if !scanner.Scan() {
			return "", false
		}
		return scanner.Text(), true
	}

	agent := NewAgent(client, getUserMessage, *verbose)
	if err := agent.Run(context.TODO()); err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}
}

func NewAgent(client *genai.Client, getUserMessage func() (string, bool), verbose bool) *Agent {
	return &Agent{
		client:         client,
		getUserMessage: getUserMessage,
		verbose:        verbose,
	}
}

type Agent struct {
	client         *genai.Client
	getUserMessage func() (string, bool)
	verbose        bool
}

func (a *Agent) Run(ctx context.Context) error {
	conversation := []*genai.Content{}

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

		// Skip empty messages
		if userInput == "" {
			if a.verbose {
				log.Println("Skipping empty message")
			}
			continue
		}

		if a.verbose {
			log.Printf("User input received: %q", userInput)
		}

		userMessage := []*genai.Part{}
		userMessage = append(userMessage, genai.NewPartFromText(userInput))
		conversation = append(conversation, genai.NewContentFromParts(userMessage, genai.RoleUser))

		if a.verbose {
			log.Printf("Sending message to Gemini, conversation length: %d", len(conversation))
		}

		message, err := a.runInference(ctx, conversation)
		if err != nil {
			if a.verbose {
				log.Printf("Error during inference: %v", err)
			}
			return err
		}

		agentMessage := []*genai.Part{}
		agentMessage = append(agentMessage, genai.NewPartFromText(message.Text()))
		conversation = append(conversation, genai.NewContentFromParts(agentMessage, genai.RoleModel))

		if a.verbose {
			log.Printf("Received response from Gemini with %d content blocks", len(message.Candidates[0].Content.Parts))
		}
		
		fmt.Printf("\u001b[93mGemini\u001b[0m: %s\n", message.Text())
	}

	if a.verbose {
		log.Println("Chat session ended")
	}
	return nil
}

func (a *Agent) runInference(ctx context.Context, conversation []*genai.Content) (*genai.GenerateContentResponse, error) {
	if a.verbose {
		log.Printf("Making API call to Gemini with model: %s", "gemini-2.5-flash")
	}

	message, err := a.client.Models.GenerateContent(
        ctx,
        "gemini-2.5-flash",
        conversation,
        nil,
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
