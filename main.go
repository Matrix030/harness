package main

import (
	"log"

	"github.com/Matrix030/mini_harness/agent"
	"github.com/Matrix030/mini_harness/tools"
)

func main() {
	registry := NewToolRegistry()

	registry.Register(NewTool(
		"read_file",
		"Read a file for disk. Params: 'path' (string, required)",
		tools.ReadFile,
	))

	registry.Register(NewTool(
		"write_file",
		"Write content to a file. Params: 'path' (string, required), 'content' (string, required)",
		tools.WriteFile,
	))

	registry.Register(NewTool(
		"run_bash",
		"Run a bash command. Params: 'commands' (string, required)",
		tools.RunBash,
	))

	err := agent.Run(
		"Do these two things: 1) create file a.txt with content 'file A', 2) create file b.txt with content 'file B'",
		registry,
	)
	if err != nil {
		log.Fatal(err)
	}
}
