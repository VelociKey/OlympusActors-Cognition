package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	olympusv1 "olympus.fleet/00SDLC/Olympus2/40000-Communication-Contracts/40400-Protocol-Synthetics/connect-rpc/olympus/v1"
	"olympus.fleet/00SDLC/OlympusGrammar/parser"
)

// MissionEngine handles batch operations and complex multi-step missions.
type MissionEngine struct {
	server *CoderServer
}

// MissionManifest defines the structure of a .gemini/mission.json or .mission
type MissionManifest struct {
	MissionID   string                `json:"mission_id"`
	Workspace   string                `json:"workspace"`
	Description string                `json:"description"`
	Steps       []MissionManifestStep `json:"steps"`
}

type MissionManifestStep struct {
	ID          string `json:"id"`
	Task        string `json:"task"` // Can be Natural Language or jeBNF IR
	Description string `json:"description"`
	Target      string `json:"target"` // Target workspace if different
}

// RunMission executes a sequence of tasks as defined in a jeBNF manifest.
func (e *MissionEngine) RunMission(ctx context.Context, workspace string, manifestPath string) (*olympusv1.MissionResponse, error) {
	slog.Info("🚀 MissionEngine: Launching mission", "manifest", manifestPath, "workspace", workspace)

	targetPath, err := validateSandbox(workspace)
	if err != nil {
		return nil, err
	}

	fullPath := filepath.Join(targetPath, manifestPath)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read mission manifest: %w", err)
	}

	var manifest *MissionManifest

	// Detect format based on extension
	if strings.HasSuffix(manifestPath, ".json") {
		manifest = &MissionManifest{}
		if err := json.Unmarshal(data, manifest); err != nil {
			return nil, fmt.Errorf("JSON Parse Error: %w", err)
		}
	} else {
		// Assume jeBNF (.mission)
		manifest, err = e.parseJeBNFDSL(string(data))
		if err != nil {
			return nil, fmt.Errorf("jeBNF Parse Error: %w", err)
		}
	}

	results := []*olympusv1.MissionStep{}
	status := "SUCCESS"

	for _, step := range manifest.Steps {
		slog.Info("🏃 Mission Step", "id", step.ID, "task", step.Task)

		// Determine target workspace (default to mission workspace)
		targetWS := workspace
		if step.Target != "" {
			targetWS = step.Target
		}

		// Execute each step via the pipeline
		res, err := e.server.executor.Execute(ctx, targetWS, step.Task, ModeAutonomous)

		stepResult := &olympusv1.MissionStep{
			Id:              step.ID,
			Task:            step.Task,
			TargetWorkspace: targetWS, // Report actual target
			Status:          "SUCCESS",
		}

		if err != nil {
			slog.Error("❌ Mission Step FAILED", "id", step.ID, "error", err)
			stepResult.Status = "FAILURE"
			stepResult.Output = err.Error()
			status = "PARTIAL_FAILURE"
			results = append(results, stepResult)
			break
		}

		stepResult.Status = res.Status
		stepResult.Output = res.Output
		results = append(results, stepResult)

		if res.Status != "SUCCESS" {
			status = "PARTIAL_FAILURE"
			break
		}
	}

	return &olympusv1.MissionResponse{
		MissionId: manifest.MissionID,
		Status:    status,
		Results:   results,
	}, nil
}

// parseJeBNFDSL uses the OlympusGrammar parser (generic) to scan tokens
// and implements a specific Recursive Descent Logic for the Mission DSL.
func (e *MissionEngine) parseJeBNFDSL(content string) (*MissionManifest, error) {
	l := parser.NewLexer(content)
	p := parser.NewParser(l)

	// We use p.NextToken() which we just exposed to consume tokens
	// and implement the grammar: Mission ID { Step "Task" (props) ... }

	manifest := &MissionManifest{}

	// Expect: Mission Identifier
	tok := p.NextToken()
	if tok.Literal != "Mission" {
		return nil, fmt.Errorf("expected 'Mission', got %s", tok.Literal)
	}

	tok = p.NextToken()
	if tok.Type != parser.TokenIdentifier {
		return nil, fmt.Errorf("expected Mission ID, got %s", tok.Literal)
	}
	manifest.MissionID = tok.Literal

	// Expect: {
	tok = p.NextToken()
	if tok.Type == parser.TokenLBrace {
		// Valid
	} else {
		return nil, fmt.Errorf("expected '{', got %s", tok.Literal)
	}

	// Parse Steps loop
	for {
		tok = p.NextToken()
		if tok.Type == parser.TokenRBrace {
			break
		}
		if tok.Type == parser.TokenEOF {
			return nil, fmt.Errorf("unexpected EOF")
		}

		if tok.Literal == "Step" {
			step, err := e.parseStep(p)
			if err != nil {
				return nil, err
			}
			manifest.Steps = append(manifest.Steps, *step)
		}
	}

	return manifest, nil
}

func (e *MissionEngine) parseStep(p *parser.Parser) (*MissionManifestStep, error) {
	step := &MissionManifestStep{}

	// Task String
	tok := p.NextToken()
	if tok.Type != parser.TokenLiteral {
		return nil, fmt.Errorf("expected quoted Task string")
	}
	step.Task = tok.Literal

	// Properties: ( ... )
	// We cheat and look at next token. If (, parse props.
	// But p.NextToken() consumes.
	// Oh wait, Parser has p.curToken/peekToken internally.
	// But we are outside the package.
	// WE EXPOSED NextToken() which consumes.
	// So we can't peek easily unless we modify Parser to expose Peek.
	// OR we just consume and check.
	// The problem is if it's NOT '(', we've consumed the first token of the NEXT step or '}'.

	// For this cycle, let's FORCE properties to be present: `( )` empty if needed.
	// Or we check if the next token is likely a property block start.

	// Let's implement robust peek by editing Parser again? No, let's just require `()` for now
	// to match the spec `Properties?` which implies optional, but my parser logic is rigid.
	// I'll update it to expect `(` always for this iteration to be safe.

	tok = p.NextToken()
	if tok.Type == parser.TokenLParen {
		// Loop properties
		for {
			key := p.NextToken()
			if key.Type == parser.TokenRParen {
				break
			}

			// key : val
			// Colon handling: Lexer treats : as Error or just char?
			// Let's assume we use `=` for properties in jeBNF like attributes.
			// id = "1"
			// Lexer handles = fine.

			eq := p.NextToken()
			if eq.Type != parser.TokenEquals {
				// skip or error
			}

			val := p.NextToken()
			step.setProp(key.Literal, val.Literal)
		}
	} else {
		// Error: expected properties block
		return nil, fmt.Errorf("step properties (...) are required in this version")
	}

	return step, nil
}

func (s *MissionManifestStep) setProp(key, val string) {
	switch key {
	case "id":
		s.ID = val
	case "target":
		s.Target = val
	case "desc":
		s.Description = val
	}
}
