package main

import (
	"context"
	"fmt"
	"log/slog"

	"connectrpc.com/connect"
	olympusv1 "olympus.fleet/00SDLC/OlympusFabric/40000-Communication-Contracts/40400-Protocol-Synthetics/connect-rpc/v1/agent"
	"olympus.fleet/00SDLC/OlympusFabric/40000-Communication-Contracts/40400-Protocol-Synthetics/connect-rpc/v1/agent/agentv1connect"
)

// MissionPlanner orchestrates the generation of new missions based on history.
type MissionPlanner struct {
	memory       agentv1connect.MemoryServiceClient
	intelligence agentv1connect.InferenceServiceClient
}

// ProposeNextMission synthesizes history and strategy to generate a mission candiate.
func (p *MissionPlanner) ProposeNextMission(ctx context.Context, objective string) (*olympusv1.SynthesisResponse, error) {
	// 1. Synthesize History
	slog.Info("Synthesizing context for mission proposal", "objective", objective)
	histResp, err := p.memory.SynthesizeContext(ctx, connect.NewRequest(&olympusv1.SynthesisRequest{
		Objective:    objective,
		ContextDepth: 10,
		Style:        "Kidder",
	}))
	if err != nil {
		return nil, fmt.Errorf("failed to synthesize history: %w", err)
	}

	// 2. Generate Proposal
	prompt := fmt.Sprintf("Based on the historical context provided below, propose the next logical mission for the Olympus Fleet. Use the 'Mission' jeBNF grammar. CONTEXT: %s", histResp.Msg.Summary)

	proposal, err := p.intelligence.Reason(ctx, connect.NewRequest(&olympusv1.ReasonRequest{
		Prompt: prompt,
		Model:  "gemini-2.0-flash",
	}))
	if err != nil {
		return nil, fmt.Errorf("proposal reasoning failed: %w", err)
	}

	return &olympusv1.SynthesisResponse{
		Summary: proposal.Msg.Output,
	}, nil
}
