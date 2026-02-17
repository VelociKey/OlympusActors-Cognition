package main

import (
	"fmt"
	"strings"
)

// MutationTask represents the IR of a mutation task (MTG)
// derived from the jeBNF mutation_task.ebnf grammar.
type MutationTask struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Mutation    MutationOp `json:"mutation"`
	Validations []string   `json:"validations,omitempty"`
}

type MutationOp struct {
	Type    string `json:"type"` // rename, replace, inject, cleanup
	Old     string `json:"old,omitempty"`
	New     string `json:"new,omitempty"`
	Pattern string `json:"pattern,omitempty"`
	Value   string `json:"value,omitempty"`
	Snippet string `json:"snippet,omitempty"`
	Anchor  string `json:"anchor,omitempty"`
	Scope   string `json:"scope,omitempty"`
	In      string `json:"in,omitempty"` // scope path
}

// AssemblePrompt translates the MutationTask IR into a high-fidelity Gemini prompt.
func AssemblePrompt(mt MutationTask) string {
	var sb strings.Builder
	sb.WriteString("ACT AS AN AUTONOMOUS CODER WITHIN THE OLYMPUS ECOSYSTEM.\n")
	sb.WriteString("EXECUTE THE FOLLOWING STRUCTURED CODE MUTATION.\n\n")

	sb.WriteString(fmt.Sprintf("MISSION: %s\n", mt.Name))
	sb.WriteString(fmt.Sprintf("OBJECTIVE: %s\n", mt.Description))

	sb.WriteString("\n--- MUTATION INSTRUCTIONS ---\n")
	switch mt.Mutation.Type {
	case "rename":
		sb.WriteString(fmt.Sprintf("ACTION: Rename '%s' to '%s'.\n", mt.Mutation.Old, mt.Mutation.New))
	case "replace":
		sb.WriteString(fmt.Sprintf("ACTION: Replace exactly '%s' with '%s'.\n", mt.Mutation.Pattern, mt.Mutation.Value))
	case "inject":
		sb.WriteString(fmt.Sprintf("ACTION: Inject the following snippet:\n```\n%s\n```\nPOSITION: %s\n", mt.Mutation.Snippet, mt.Mutation.Anchor))
	case "cleanup":
		sb.WriteString(fmt.Sprintf("ACTION: Perform nomenclature cleanup based on standard: %s.\n", mt.Mutation.Old))
	}

	if mt.Mutation.Scope != "" {
		sb.WriteString(fmt.Sprintf("SCOPE: %s\n", mt.Mutation.Scope))
	}

	sb.WriteString("\n--- GOVERNANCE & QUALITY ---\n")
	sb.WriteString("1. MANDATORY: Follow ADR-010 for naming (Clean functional names, no 'Agent' suffix unless specified).\n")
	sb.WriteString("2. SAFETY: Only modify string literals, log messages, and pulse responses unless structural change is requested.\n")
	sb.WriteString("3. INTEGRITY: Do not break existing imports or circular dependencies.\n")

	if len(mt.Validations) > 0 {
		sb.WriteString("\n--- POST-CONDITION CHECKS ---\n")
		for _, v := range mt.Validations {
			sb.WriteString(fmt.Sprintf("- Run: %s\n", v))
		}
	}

	return sb.String()
}
