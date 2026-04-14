package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
)

const ollamaURL = "http://localhost:11434/api/chat"
const model = "gemma4:26b"

// ---- Request / Response shapes Ollama expects ----

type Message struct {
	Role      string     `json:"role"`
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

type ToolCall struct {
	Function ToolCallFunction `json:"function"`
}

type ToolCallFunction struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

type OllamaRequest struct {
	Model    string     `json:"model"`
	Messages []Message  `json:"messages"`
	Tools    []any      `json:"tools"`
	Stream   bool       `json:"stream"`
	Options  OllamaOpts `json:"options"`
}

type OllamaOpts struct {
	Think bool `json:"think"`
}

type OllamaResponse struct {
	Message Message `json:"message"`
}

// ---- The Runner interface - anything with Run() can be used ----

type Runner interface {
	Run(name string, params map[string]any) (string, error)
	AsOllamaTools() []any
}

func Run(goal string, registry Runner) error {
	messages := []Message{
		{Role: "user", Content: goal},
	}

	fmt.Printf("\nGoal: %s\n\n", goal)

	for {
		// 1. Build and send request to Ollama
		reqBody, err := json.Marshal(OllamaRequest{
			Model:    model,
			Messages: messages,
			Tools:    registry.AsOllamaTools(),
			Stream:   false,
			Options:  OllamaOpts{Think: false},
		})

		if err != nil {
			return fmt.Errorf("marshal error: %w", err)
		}

		resp, err := http.Post(ollamaURL, "application/json", bytes.NewReader(reqBody))
		if err != nil {
			return fmt.Errorf("ollama request failed: %w", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		var ollamaResp OllamaResponse
		if err := json.Unmarshal(body, &ollamaResp); err != nil {
			return fmt.Errorf("unmarshal error: %w", err)
		}

		msg := ollamaResp.Message
		messages = append(messages, msg)

		// 2. Tool calls requested?
		if len(msg.ToolCalls) > 0 {
			type toolResult struct {
				content string
			}

			results := make([]toolResult, len(msg.ToolCalls))
			var wg sync.WaitGroup
			var mu sync.Mutex

			for i, call := range msg.ToolCalls {
				wg.Add(1)

				// capture loop vars - critical in Go goroutines
				i, call := i, call

				go func() {
					defer wg.Done()

					name := call.Function.Name
					params := call.Function.Arguments

					fmt.Printf("Tool: %s | Params: %v\n", name, params)

					output, err := registry.Run(name, params)
					if err != nil {
						output = fmt.Sprintf("ERROR: %s", err.Error())
					}

					fmt.Printf("Result: %.200s\n\n", output)

					mu.Lock()
					results[i] = toolResult{content: output}
					mu.Unlock()
				}()
			}

			wg.Wait() // block until all tools finish

			// feed all results back in order
			for _, r := range results {
				messages = append(messages, Message{
					Role:    "tool",
					Content: r.content,
				})
			}

		} else {
			// 3. No tool call = agent is done
			fmt.Printf("\nAgent: %s\n", msg.Content)
			return nil
		}
	}
}
