package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// SandboxClient calls the SandboxAPI HTTP API.
type SandboxClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewSandboxClient creates a client for the SandboxAPI HTTP API.
func NewSandboxClient(baseURL, apiKey string) *SandboxClient {
	return &SandboxClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
	}
}

// ExecuteRequest is the request body for POST /v1/execute.
type ExecuteRequest struct {
	Language string `json:"language"`
	Code     string `json:"code"`
	Timeout  int    `json:"timeout,omitempty"`
	Stdin    string `json:"stdin,omitempty"`
}

// ExecuteResult is the response from POST /v1/execute.
type ExecuteResult struct {
	ID              string `json:"id"`
	Status          string `json:"status"`
	Language        string `json:"language"`
	Stdout          string `json:"stdout"`
	Stderr          string `json:"stderr"`
	ExitCode        int    `json:"exit_code"`
	ExecutionTimeMs int64  `json:"execution_time_ms"`
	MemoryUsedKB    int64  `json:"memory_used_kb"`
}

// LanguageInfo is the response shape for each language from GET /v1/languages.
type LanguageInfo struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Version       string   `json:"version"`
	Aliases       []string `json:"aliases"`
	FileExtension string   `json:"file_extension"`
	Example       string   `json:"example"`
}

// LanguageListResponse is the response from GET /v1/languages.
type LanguageListResponse struct {
	Languages []LanguageInfo `json:"languages"`
}

// APIError represents an error response from the API.
type APIError struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// Execute sends code to the SandboxAPI for execution.
func (c *SandboxClient) Execute(ctx context.Context, req ExecuteRequest) (*ExecuteResult, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/v1/execute", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-RapidAPI-Proxy-Secret", c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var apiErr APIError
		if json.Unmarshal(respBody, &apiErr) == nil && apiErr.Error != "" {
			return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, apiErr.Error)
		}
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(respBody))
	}

	var result ExecuteResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &result, nil
}

// ListLanguages fetches the list of supported languages from the API.
func (c *SandboxClient) ListLanguages(ctx context.Context) ([]LanguageInfo, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.baseURL+"/v1/languages", nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("X-RapidAPI-Proxy-Secret", c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("list languages request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(respBody))
	}

	var listResp LanguageListResponse
	if err := json.Unmarshal(respBody, &listResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return listResp.Languages, nil
}
