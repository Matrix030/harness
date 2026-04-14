package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Tools    []any     `json:"tools"`
	Stream   bool      `json:"stream"`
}

type OllamaResponse struct {
	Message Message `json:"message"`
}

// ----- The Runner interface - anything with Run() can be used ----

type Runner interface {
	Run(name string, params map[string]any) (string, error)
	AsOllamaTools() []any
}

func Run(goal string, registry Runner) error {
	messages := []Message{
		{Role: "user", Content: goal},
	}

	fmt.Printf("\n Goal: %s\n\n", goal)

	for {
		// 1. Build and send request to Ollama
		reqBody, err := json.Marshal(OllamaRequest{
			Model:    model,
			Messages: messages,
			Tools:    registry.AsOllamaTools(),
			Stream:   false,
		})

		if err != nil {
			return fmt.Errorf("marshall error: %w", err)
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

		// 2. Tool call requested?
		if len(msg.ToolCalls) > 0 {
			for _, call := range msg.ToolCalls {
				name := call.Function.Name
				params := call.Function.Arguments

				fmt.Printf("Tool: %s | Params: %v\n", name, params)

				result, err := registry.Run(name, params)
				if err != nil {
					result = fmt.Sprintf("ERROR: %s", err.Error())
				}

				fmt.Printf("Result: %.200\n\n", result) // cap at 200 chars

				// 3. Feed result back
				messages = append(messages, Message{
					Role:    "tool",
					Content: result,
				})
			}
		} else {
			// 4. No tool call = agent is done
			fmt.Printf("Agent: %s\n", msg.Content)
			return nil
		}
	}
}
