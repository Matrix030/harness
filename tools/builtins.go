package tools

import (
	"fmt"
	"os"
	"os/exec"
)

func ReadFile(params map[string]any) (string, error) {
	path, ok := params["path"].(string)
	if !ok {
		return "", fmt.Errorf("param 'path' required and must be a string")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func WriteFile(params map[string]any) (string, error) {
	path, ok := params["path"].(string)
	if !ok {
		// fallback - handle model sending 'filename'
		path, ok = params["filename"].(string)
		if !ok {
			return "", fmt.Errorf("params 'path' required and must be a string")
		}
	}
	content, ok := params["content"].(string)
	if !ok {
		return "", fmt.Errorf("param 'content' required and must be a string")
	}
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("written to %s", path), nil
}

func RunBash(params map[string]any) (string, error) {
	command, ok := params["command"].(string)
	if !ok {
		return "", fmt.Errorf("param 'command' required and must be a string")
	}
	out, err := exec.Command("bash", "-c", command).CombinedOutput()
	if err != nil {
		return string(out), err
	}
	return string(out), nil
}
