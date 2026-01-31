package llmgen

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Provider defines the interface for LLM providers.
type Provider interface {
	Generate(ctx context.Context, req GenerateRequest) (*GenerateResponse, error)
	Name() string
}

// GenerateRequest represents a generation request to an LLM.
type GenerateRequest struct {
	Prompt      string                 `json:"prompt"`
	Model       string                 `json:"model"`
	Temperature float64                `json:"temperature,omitempty"`
	MaxTokens   int                    `json:"max_tokens,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// GenerateResponse represents an LLM response.
type GenerateResponse struct {
	Content   string                 `json:"content"`
	Model     string                 `json:"model"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// OllamaProvider implements Provider for Ollama.
type OllamaProvider struct {
	BaseURL string
	Client  *http.Client
}

// NewOllamaProvider creates a new Ollama provider.
func NewOllamaProvider(baseURL string) *OllamaProvider {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	return &OllamaProvider{
		BaseURL: baseURL,
		Client:  &http.Client{Timeout: 120 * time.Second},
	}
}

// Name returns the provider name.
func (p *OllamaProvider) Name() string {
	return "ollama"
}

// Generate sends a generation request to Ollama.
func (p *OllamaProvider) Generate(ctx context.Context, req GenerateRequest) (*GenerateResponse, error) {
	payload := map[string]interface{}{
		"model":  req.Model,
		"prompt": req.Prompt,
		"stream": false,
	}

	if req.Temperature > 0 {
		payload["options"] = map[string]interface{}{
			"temperature": req.Temperature,
		}
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.BaseURL+"/api/generate", bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Response string `json:"response"`
		Model    string `json:"model"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &GenerateResponse{
		Content:   result.Response,
		Model:     result.Model,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"temperature": req.Temperature,
		},
	}, nil
}

// OpenAICompatibleProvider implements Provider for OpenAI-compatible APIs.
type OpenAICompatibleProvider struct {
	BaseURL string
	APIKey  string
	Client  *http.Client
}

// NewOpenAICompatibleProvider creates a new OpenAI-compatible provider.
func NewOpenAICompatibleProvider(baseURL, apiKey string) *OpenAICompatibleProvider {
	return &OpenAICompatibleProvider{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Client:  &http.Client{Timeout: 120 * time.Second},
	}
}

// Name returns the provider name.
func (p *OpenAICompatibleProvider) Name() string {
	return "openai-compatible"
}

// Generate sends a generation request to the OpenAI-compatible API.
func (p *OpenAICompatibleProvider) Generate(ctx context.Context, req GenerateRequest) (*GenerateResponse, error) {
	payload := map[string]interface{}{
		"model": req.Model,
		"messages": []map[string]string{
			{"role": "user", "content": req.Prompt},
		},
	}

	if req.Temperature > 0 {
		payload["temperature"] = req.Temperature
	}
	if req.MaxTokens > 0 {
		payload["max_tokens"] = req.MaxTokens
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.BaseURL+"/v1/chat/completions", bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if p.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+p.APIKey)
	}

	resp, err := p.Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("api returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Model string `json:"model"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	return &GenerateResponse{
		Content:   result.Choices[0].Message.Content,
		Model:     result.Model,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"temperature": req.Temperature,
			"max_tokens":  req.MaxTokens,
		},
	}, nil
}
