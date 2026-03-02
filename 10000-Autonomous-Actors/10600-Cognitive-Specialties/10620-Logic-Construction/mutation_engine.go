package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"olympus.fleet/00SDLC/OlympusGrammar/20000-Context-Bridges/000-olympus.fleet/00SDLC/OlympusGrammar/P0000-pkg/000-parser"
)

// MutationEngine (MGS) processes structured mutation tasks
// conforming to the Mutation Task Grammar (MTG).
type MutationEngine struct {
	server *CoderServer
}

// ExecuteMutationIR takes a raw DSL or JSON IR,
// translates it via the Prompt Factory, and sends it to Gemini.
func (e *MutationEngine) ExecuteMutationIR(ctx context.Context, workspace string, input string, dryRun bool) (string, error) {
	var mt MutationTask
	var err error

	if strings.HasPrefix(strings.TrimSpace(input), "{") {
		if err := json.Unmarshal([]byte(input), &mt); err != nil {
			return "", fmt.Errorf("MTG Syntax Error (Invalid JSON IR): %w", err)
		}
	} else {
		parsedMT, err := e.parseJeBNF(input)
		if err != nil {
			return "", fmt.Errorf("MTG Syntax Error (Invalid jeBNF): %w", err)
		}
		mt = *parsedMT
	}

	slog.Info("🧪 MutationEngine: Processing blueprint", "blueprint", mt.Name, "workspace", workspace)

	// Step 1: Translate IR Node to High-Fidelity Prompt
	prompt := AssemblePrompt(mt)

	// Step 2: Validate Sandbox
	targetPath, err := validateSandbox(workspace)
	if err != nil {
		return "", err
	}

	// Step 3: Dispatch to Gemini CLI via the existing RunGemini method
	// We pass the context to respect timeouts/cancellation
	return e.server.RunGemini(ctx, targetPath, prompt, dryRun)
}

func (e *MutationEngine) parseJeBNF(content string) (*MutationTask, error) {
	l := parser.NewLexer(content)
	p := parser.NewParser(l)

	mt := &MutationTask{}

	// Expect: name : identifier
	tok := p.NextToken()
	if tok.Literal != "name" {
		return nil, fmt.Errorf("expected 'name', got %s", tok.Literal)
	}
	if p.NextToken().Type != parser.TokenColon {
		return nil, fmt.Errorf("expected ':' after name")
	}
	mt.Name = p.NextToken().Literal

	// Expect: description : string
	tok = p.NextToken()
	if tok.Literal != "description" {
		return nil, fmt.Errorf("expected 'description', got %s", tok.Literal)
	}
	if p.NextToken().Type != parser.TokenColon {
		return nil, fmt.Errorf("expected ':' after description")
	}
	mt.Description = p.NextToken().Literal

	// Expect Mutation Op: rename | replace | inject | cleanup
	tok = p.NextToken()
	mt.Mutation.Type = tok.Literal
	switch tok.Literal {
	case "rename":
		mt.Mutation.Old = p.NextToken().Literal // identifier
		if p.NextToken().Literal != "to" {
			return nil, fmt.Errorf("expected 'to' in rename")
		}
		mt.Mutation.New = p.NextToken().Literal
	case "replace":
		mt.Mutation.Pattern = p.NextToken().Literal // string
		if p.NextToken().Literal != "with" {
			return nil, fmt.Errorf("expected 'with' in replace")
		}
		mt.Mutation.Value = p.NextToken().Literal
	case "inject":
		mt.Mutation.Snippet = p.NextToken().Literal // string
		if p.NextToken().Literal != "at" {
			return nil, fmt.Errorf("expected 'at' in inject")
		}
		mt.Mutation.Anchor = p.NextToken().Literal
	case "cleanup":
		if p.NextToken().Literal != "nomenclature" {
			return nil, fmt.Errorf("expected 'nomenclature' after cleanup")
		}
		if p.NextToken().Literal != "based" {
			// Skip 'based on'
		}
		if p.NextToken().Literal != "on" {
			// Skip
		}
		mt.Mutation.Old = p.NextToken().Literal // identifier
	}

	// Optional scope: in ...
	tok = p.NextToken()
	if tok.Literal == "in" {
		mt.Mutation.Scope = p.NextToken().Literal // file | directory | workspace
		mt.Mutation.In = p.NextToken().Literal    // path
	}

	return mt, nil
}
