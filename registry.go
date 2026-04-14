package main

import "fmt"

type ToolRegistry struct {
	tools map[string]Tool
}

func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]Tool),
	}
}

func (r *ToolRegistry) Register(t Tool) {
	r.tools[t.Name()] = t
}

func (r *ToolRegistry) Run(name string, params map[string]any) (string, error) {
	t, ok := r.tools[name]
	if !ok {
		return "", fmt.Errorf("unknown tool: %s", name)
	}
	return t.Run(params)
}

// OllamaToolDef is the shape Ollama expects for each tool
type OllamaToolDef struct {
	Type     string         `json:"type"`
	Function OllamaFunction `json:"function"`
}

type OllamaFunction struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (r *ToolRegistry) AsOllamaTools() []OllamaToolDef {
	defs := make([]OllamaToolDef, 0, len(r.tools))
	for _, t := range r.tools {
		defs = append(defs, OllamaToolDef{
			Type: "function",
			Function: OllamaFunction{
				Name:        t.Name(),
				Description: t.Description(),
			},
		})
	}
	return defs
}
