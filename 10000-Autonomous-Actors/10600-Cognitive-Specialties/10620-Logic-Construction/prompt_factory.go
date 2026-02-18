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
	sb.WriteString(fmt.Sprintf("OBJECTIVE: %s\n\n", mt.Description))

	sb.WriteString("INSTRUCTIONS:\n")
	switch mt.Mutation.Type {
	case "rename":
		sb.WriteString(fmt.Sprintf("- Rename symbol '%s' to '%s' in scope '%s'.\n", mt.Mutation.Old, mt.Mutation.New, mt.Mutation.In))
	case "replace":
		sb.WriteString(fmt.Sprintf("- Replace pattern '%s' with value '%s' in file '%s'.\n", mt.Mutation.Pattern, mt.Mutation.Value, mt.Mutation.In))
	case "inject":
		sb.WriteString(fmt.Sprintf("- Inject the following snippet near anchor '%s' in file '%s':\n%s\n", mt.Mutation.Anchor, mt.Mutation.In, mt.Mutation.Snippet))
	case "cleanup":
		sb.WriteString(fmt.Sprintf("- Perform logical cleanup in scope '%s'.\n", mt.Mutation.In))
	}

	if len(mt.Validations) > 0 {
		sb.WriteString("\nPOST-MUTATION VALIDATIONS:\n")
		for _, v := range mt.Validations {
			sb.WriteString(fmt.Sprintf("- %s\n", v))
		}
	}

	sb.WriteString("\nRESPOND ONLY WITH THE FINAL MODIFIED CODE OR THE ACTIONS PERFORMED. DO NOT ADD PREAMBLES.")
	return sb.String()
}
