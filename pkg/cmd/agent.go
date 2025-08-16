package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/genai"

	"daviderutigliano/agentic-ai/internal/agent"
	"daviderutigliano/agentic-ai/internal/tools"
)

func main() {
	var verbose bool
	var toolNames []string

	rootCmd := &cobra.Command{
		Use:   "agent",
		Short: "A chat agent powered by Gemini.",
		Run: func(cmd *cobra.Command, args []string) {
			run(cmd.Context(), verbose, toolNames)
		},
	}

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose logging")
	rootCmd.PersistentFlags().StringSliceVar(&toolNames, "tools", []string{}, "a list of tools to use")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(ctx context.Context, verbose bool, toolNames []string) {
	if verbose {
		log.SetOutput(os.Stderr)
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Println("Verbose logging enabled")
	} else {
		log.SetOutput(os.Stdout)
		log.SetFlags(0)
		log.SetPrefix("")
	}

	// The client gets the API key from the environment variable `GEMINI_API_KEY`.
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	var activeTools []*tools.ToolDefinition
	if len(toolNames) > 0 {
		if verbose {
			log.Printf("Enabling tools: %s", strings.Join(toolNames, ", "))
		}
		for _, name := range toolNames {
			tool, ok := tools.AvailableTools[name]
			if !ok {
				log.Fatalf("Unknown tool: %s", name)
			}
			activeTools = append(activeTools, tool)
		}
	}

	scanner := bufio.NewScanner(os.Stdin)
	getUserMessage := func() (string, bool) {
		if !scanner.Scan() {
			return "", false
		}
		return scanner.Text(), true
	}

	a := agent.NewAgent(client, getUserMessage, activeTools, verbose)
	if err := a.Run(ctx); err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}
}
