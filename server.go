package main

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// NewMCPServer creates a configured MCP server with all tools registered.
func NewMCPServer(version string, client *SandboxClient) *server.MCPServer {
	s := server.NewMCPServer(
		"sandboxapi",
		version,
		server.WithToolCapabilities(true),
	)

	tools := &ToolHandlers{client: client}

	s.AddTool(executeCodeTool(), tools.ExecuteCode)
	s.AddTool(executeBatchTool(), tools.ExecuteBatch)
	s.AddTool(listLanguagesTool(), tools.ListLanguages)

	return s
}

// executeCodeTool defines the execute_code tool schema.
func executeCodeTool() mcp.Tool {
	return mcp.NewTool("execute_code",
		mcp.WithDescription("Execute code in a sandboxed container. Supports Python, JavaScript, TypeScript, Bash, Java, C++, C, and Go."),
		mcp.WithString("language",
			mcp.Required(),
			mcp.Description("Programming language to execute"),
			mcp.Enum("python3", "javascript", "typescript", "bash", "java", "cpp", "c", "go"),
		),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("Source code to execute"),
		),
		mcp.WithNumber("timeout",
			mcp.Description("Execution timeout in seconds (default 10, max 300)"),
		),
		mcp.WithString("stdin",
			mcp.Description("Standard input to pass to the program"),
		),
	)
}

// executeBatchTool defines the execute_batch tool schema.
func executeBatchTool() mcp.Tool {
	return mcp.NewTool("execute_batch",
		mcp.WithDescription("Execute multiple code snippets in sandboxed containers. Each execution runs independently."),
		mcp.WithArray("executions",
			mcp.Required(),
			mcp.Description("Array of execution requests, each with language, code, and optional timeout/stdin"),
		),
	)
}

// listLanguagesTool defines the list_languages tool schema.
func listLanguagesTool() mcp.Tool {
	return mcp.NewTool("list_languages",
		mcp.WithDescription("List all supported programming languages with their versions and examples"),
	)
}
