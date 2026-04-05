package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// ToolHandlers holds the HTTP client used by MCP tool handlers.
type ToolHandlers struct {
	client *SandboxClient
}

// executeCodeResponse is the JSON shape returned by execute_code.
type executeCodeResponse struct {
	ID              string `json:"id"`
	Status          string `json:"status"`
	Language        string `json:"language"`
	Stdout          string `json:"stdout"`
	Stderr          string `json:"stderr"`
	ExitCode        int    `json:"exit_code"`
	ExecutionTimeMs int64  `json:"execution_time_ms"`
	MemoryUsedKB    int64  `json:"memory_used_kb"`
}

// batchResponse is the JSON shape returned by execute_batch.
type batchResponse struct {
	Results   []executeCodeResponse `json:"results"`
	Total     int                   `json:"total"`
	Completed int                   `json:"completed"`
	Failed    int                   `json:"failed"`
}

// ExecuteCode handles the execute_code tool call.
func (h *ToolHandlers) ExecuteCode(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	lang, err := request.RequireString("language")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid parameters: %s", err)), nil
	}

	code, err := request.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid parameters: %s", err)), nil
	}

	timeout := request.GetInt("timeout", 10)
	stdin := request.GetString("stdin", "")

	result, err := h.client.Execute(ctx, ExecuteRequest{
		Language: lang,
		Code:     code,
		Timeout:  timeout,
		Stdin:    stdin,
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("execution failed: %s", err)), nil
	}

	resp := executeCodeResponse{
		ID:              result.ID,
		Status:          result.Status,
		Language:        result.Language,
		Stdout:          result.Stdout,
		Stderr:          result.Stderr,
		ExitCode:        result.ExitCode,
		ExecutionTimeMs: result.ExecutionTimeMs,
		MemoryUsedKB:    result.MemoryUsedKB,
	}

	jsonBytes, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("marshal result: %w", err)
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
}

// ExecuteBatch handles the execute_batch tool call.
func (h *ToolHandlers) ExecuteBatch(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	execsRaw, ok := args["executions"]
	if !ok {
		return mcp.NewToolResultError("missing required parameter: executions"), nil
	}

	rawJSON, err := json.Marshal(execsRaw)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid executions parameter: %s", err)), nil
	}

	var executions []ExecuteRequest
	if err := json.Unmarshal(rawJSON, &executions); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid executions format: %s", err)), nil
	}

	if len(executions) == 0 {
		return mcp.NewToolResultError("executions array must not be empty"), nil
	}

	results := make([]executeCodeResponse, 0, len(executions))
	completed := 0
	failed := 0

	for _, exec := range executions {
		if exec.Timeout <= 0 {
			exec.Timeout = 10
		}

		result, err := h.client.Execute(ctx, exec)
		if err != nil {
			results = append(results, executeCodeResponse{
				Status:   "error",
				Language: exec.Language,
				Stderr:   fmt.Sprintf("execution failed: %s", err),
				ExitCode: -1,
			})
			failed++
			continue
		}

		results = append(results, executeCodeResponse{
			ID:              result.ID,
			Status:          result.Status,
			Language:        result.Language,
			Stdout:          result.Stdout,
			Stderr:          result.Stderr,
			ExitCode:        result.ExitCode,
			ExecutionTimeMs: result.ExecutionTimeMs,
			MemoryUsedKB:    result.MemoryUsedKB,
		})

		if result.ExitCode == 0 {
			completed++
		} else {
			failed++
		}
	}

	batch := batchResponse{
		Results:   results,
		Total:     len(executions),
		Completed: completed,
		Failed:    failed,
	}

	jsonBytes, err := json.Marshal(batch)
	if err != nil {
		return nil, fmt.Errorf("marshal batch result: %w", err)
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
}

// ListLanguages handles the list_languages tool call.
func (h *ToolHandlers) ListLanguages(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	langs, err := h.client.ListLanguages(ctx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list languages: %s", err)), nil
	}

	jsonBytes, err := json.Marshal(langs)
	if err != nil {
		return nil, fmt.Errorf("marshal languages: %w", err)
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
}
