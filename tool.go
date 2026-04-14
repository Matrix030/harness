package main

// Tool is an interface - any struct that implements Run() is a Tool.
type Tool interface {
	Name() string
	Description() string
	Run(params map[string]any) (string, error)
}

// BaseTool is a concrete struct that satisfies the Tool interface
type BaseTool struct {
	name        string
	description string
	fn          func(map[string]any) (string, error)
}

func NewTool(name, description string, fn func(map[string]any) (string, error)) *BaseTool {
	return &BaseTool{
		name:        name,
		description: description,
		fn:          fn,
	}
}

// These three methods make BaseTool satisfy the Tool interface
func (t *BaseTool) Name() string        { return t.name }
func (t *BaseTool) Description() string { return t.description }
func (t *BaseTool) Run(params map[string]any) (string, error) {
	return t.fn(params)
}
