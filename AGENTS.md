# Agent Instructions

## Commands
- `go build` - Build Go project
- `go run cms/agent.go` - Run the chat application
- `go test ./...` - Run all Go tests
- `go test <package>` - Run tests for specific package
- `go mod tidy` - Download dependencies

### Application Commands
- `go run agent.go` - Simple chat interface with Gemini
- `go run agent.go --tools read_file` - Chat with file reading capabilities

### Verbose Logging
All Go applications support a `--verbose` flag for detailed execution logging:
- `go run agent.go [--tools <tool1>, <tool2>, ...] --verbose` - Enable verbose logging for debugging

## Architecture
- **Environment**: Nix-based development environment using devenv
- **Shell**: Includes Git, Go toolchain, and custom greeting script
- **Structure**: Chat application with terminal interface to Gemini via Gemini API

## Troubleshooting

### Verbose Logging
When debugging issues with the chat applications, use the `--verbose` flag to get detailed execution logs:

```bash
go run edit_tool.go --verbose
```

**What verbose logging shows:**
- API calls to Gemini (model, timing, success/failure)
- Tool execution details (which tools are called, input parameters, results)
- File operations (reading, writing, listing files with sizes/counts)
- Bash command execution (commands run, output, errors)
- Conversation flow (message processing, content blocks)
- Error details with stack traces

**Log output locations:**
- **Verbose mode**: Detailed logs go to stderr with timestamps and file locations
- **Normal mode**: Only essential output goes to stdout

**Common troubleshooting scenarios:**
- **API failures**: Check verbose logs for authentication errors or rate limits
- **Tool failures**: See exactly which tool failed and why (file not found, permission errors)
- **Unexpected responses**: View full conversation flow and Gemini's reasoning
- **Performance issues**: See API call timing and response sizes

### Environment Issues
- Ensure `GEMINI_API_KEY` environment variable is set
- Run `devenv shell` to ensure proper development environment
- Use `go mod tidy` to ensure dependencies are installed

## Notes
- Requires GEMINI_API_KEY environment variable to be set
- Chat application provides a simple terminal interface to Gemini
- Use ctrl-c to quit the chat session