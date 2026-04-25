// KhunQuant - Ultra-lightweight personal AI agent
// Inspired by and based on nanobot: https://github.com/HKUDS/nanobot
// License: MIT
//
// Copyright (c) 2026 KhunQuant contributors

package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/cryptoquantumwave/khunquant/pkg/providers/openai_compat"
)

// MLXLMProvider wraps the OpenAI-compatible HTTP provider for mlx_lm servers.
// Because mlx_lm loads exactly one model at startup and uses the model field
// to decide whether to download a new model, we auto-discover the loaded model
// ID via GET /v1/models and use that in every chat request.
//
// If configuredModel is non-empty, it is used as-is (no auto-discovery). This
// is needed when the server is loaded from a local directory path: newer
// mlx_lm versions return the HuggingFace repo ID from GET /v1/models rather
// than the local path, causing a mismatch that triggers a download. Set
// configuredModel to match the --model argument passed to the mlx_lm server.
type MLXLMProvider struct {
	delegate        *openai_compat.Provider
	apiBase         string
	proxy           string
	configuredModel string // if set, skip auto-discovery

	mu              sync.Mutex
	discoveredModel string
}

func NewMLXLMProvider(apiKey, apiBase, proxy string, requestTimeoutSeconds int, configuredModel string) *MLXLMProvider {
	return &MLXLMProvider{
		delegate: openai_compat.NewProvider(
			apiKey,
			apiBase,
			proxy,
			openai_compat.WithRequestTimeout(time.Duration(requestTimeoutSeconds)*time.Second),
		),
		apiBase:         apiBase,
		proxy:           proxy,
		configuredModel: configuredModel,
	}
}

// expandTilde expands a leading ~ to the user's home directory.
func expandTilde(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return filepath.Join(home, strings.TrimPrefix(path, "~"))
}

// resolveModel returns the model string to use in chat requests.
// If configuredModel is set, it is returned immediately (no auto-discovery).
// Leading ~ is expanded so the path matches what the mlx_lm server received
// after shell expansion of its --model argument.
// Otherwise, GET /v1/models is queried once and the result is cached.
func (p *MLXLMProvider) resolveModel(ctx context.Context) string {
	if p.configuredModel != "" {
		return expandTilde(p.configuredModel)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.discoveredModel != "" {
		return p.discoveredModel
	}

	model := p.fetchLoadedModel(ctx)
	if model != "" {
		p.discoveredModel = model
	}
	return model
}

func (p *MLXLMProvider) fetchLoadedModel(ctx context.Context) string {
	apiBase := strings.TrimRight(strings.TrimSpace(p.apiBase), "/")
	if apiBase == "" {
		apiBase = "http://localhost:8080/v1"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiBase+"/models", nil)
	if err != nil {
		return ""
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ""
	}

	if len(result.Data) > 0 {
		return result.Data[0].ID
	}
	return ""
}

func (p *MLXLMProvider) Chat(
	ctx context.Context,
	messages []Message,
	tools []ToolDefinition,
	model string,
	options map[string]any,
) (*LLMResponse, error) {
	resolved := p.resolveModel(ctx)
	if resolved == "" {
		return nil, fmt.Errorf("mlx_lm: could not determine loaded model (is the server running?)")
	}
	return p.delegate.Chat(ctx, messages, tools, resolved, options)
}

func (p *MLXLMProvider) GetDefaultModel() string {
	return ""
}
