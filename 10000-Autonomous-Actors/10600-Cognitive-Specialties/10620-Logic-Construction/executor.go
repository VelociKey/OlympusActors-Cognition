package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

// RiskLevel defines the potential impact of a code mutation.
type RiskLevel int

const (
	RiskLow    RiskLevel = 1 // String replacements, comments (Always Autonomous)
	RiskMedium RiskLevel = 2 // Type renames, imports (Usually Autonomous)
	RiskHigh   RiskLevel = 3 // Function signatures, logic changes (Requires Supervision)
)

// ExecutionMode defines how the pipeline handles approvals.
type ExecutionMode string

const (
	ModeAutonomous ExecutionMode = "autonomous"
	ModeSupervised ExecutionMode = "supervised"
	ModeDryRun     ExecutionMode = "dry-run"
)

// Executor manages the multi-stage execution pipeline.
type Executor struct {
	server *CoderServer
}

// PipelineResult captures the outcome of a single pipeline execution.
type PipelineResult struct {
	TaskID   string
	Status   string
	Efficacy float32
	Output   string
	Diff     string
	Report   string
}

// Execute handles the end-to-end mutation lifecycle.
func (e *Executor) Execute(ctx context.Context, workspace string, taskIR string, mode ExecutionMode) (*PipelineResult, error) {
	slog.Info("🚀 Executor: Starting pipeline", "workspace", workspace, "mode", mode)

	// Stage 1: Parse & Risk Assessment
	var mt MutationTask
	if strings.HasPrefix(strings.TrimSpace(taskIR), "{") {
		if err := json.Unmarshal([]byte(taskIR), &mt); err != nil {
			return nil, fmt.Errorf("failed to parse task JSON IR for risk assessment: %w", err)
		}
	} else {
		parsedMT, err := e.server.engine.parseJeBNF(taskIR)
		if err != nil {
			return nil, fmt.Errorf("failed to parse task jeBNF IR for risk assessment: %w", err)
		}
		mt = *parsedMT
	}

	risk := e.assessRisk(&mt)
	slog.Info("⚖️ Risk Assessment", "risk", risk, "action", mt.Mutation.Type)

	if risk == RiskHigh && mode == ModeAutonomous {
		slog.Warn("⚠️ High Risk mutation detected. Promoting to SUPERVISED mode (Audit log entry created).")
	}

	taskID := fmt.Sprintf("TASK-%d", time.Now().Unix())

	// Stage 2: Planning / Dry Run
	if mode == ModeDryRun {
		planOutput, err := e.server.engine.ExecuteMutationIR(ctx, workspace, taskIR, true)
		if err != nil {
			return nil, fmt.Errorf("planning phase failed: %w", err)
		}
		res := &PipelineResult{
			TaskID: taskID,
			Status: "PLAN_COMPLETE",
			Output: planOutput,
		}
		reporter := &Reporter{}
		res.Report = reporter.GenerateReport(res, workspace)
		return res, nil
	}

	// Stage 4: Execution & Self-Healing Loop
	maxHeals := 2
	var lastOutput string
	var lastReport *ValidationReport

	for attempt := 0; attempt <= maxHeals; attempt++ {
		// Respect context cancellation between attempts
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		if attempt > 0 {
			slog.Info("🩹 Executor: Attempting Self-Healing", "attempt", attempt, "max", maxHeals)

			// Extract error context
			var errCtx string
			if lastReport != nil {
				for step, res := range lastReport.Details {
					if res.Status == "FAIL" {
						errCtx += fmt.Sprintf("[%s Failure]\n%s\n", step, res.Output)
					}
				}
			}

			healingPrompt := fmt.Sprintf("FIX THE FOLLOWING REGRESSION FROM THE PREVIOUS ATTEMPT.\nERROR:\n%s\nORIGINAL TASK: %s", errCtx, taskIR)

			targetPath, _ := validateSandbox(workspace)
			healOut, err := e.server.RunGemini(ctx, targetPath, healingPrompt, false)
			if err != nil {
				slog.Error("❌ Executor: Healing execution failed", "attempt", attempt, "error", err)
				lastOutput = healOut
				continue
			}
			lastOutput = healOut
		} else {
			slog.Info("🔨 Executor: Executing initial mutation")
			executeOutput, err := e.server.engine.ExecuteMutationIR(ctx, workspace, taskIR, false)
			if err != nil {
				return nil, fmt.Errorf("initial execution failed: %w", err)
			}
			lastOutput = executeOutput
		}

		// Stage 5: Validate
		slog.Info("🔍 Executor: Running validation suite")
		targetPath, _ := validateSandbox(workspace)
		v := &Validator{path: targetPath}
		report, err := v.Validate(ctx)
		if err != nil {
			slog.Error("❌ Executor: Validation process failed", "error", err)
			return nil, err
		}
		lastReport = report

		if report.Success && report.Score >= 5.0 {
			slog.Info("✨ Executor: Mutation successful and verified", "attempt", attempt)
			break
		}

		if attempt == maxHeals {
			slog.Error("❌ Executor: Max healing attempts reached. Declaring failure.")
		}
	}

	status := "SUCCESS"
	if lastReport == nil || !lastReport.Success || lastReport.Score < 3.0 {
		status = "FAILURE"
	}

	// Stage 6: Report
	slog.Info("📊 Pipeline complete", "status", status, "efficacy", lastReport.Score)

	targetPath, _ := validateSandbox(workspace)
	res := &PipelineResult{
		TaskID:   taskID,
		Status:   status,
		Efficacy: float32(lastReport.Score),
		Output:   fmt.Sprintf("%s\n\nValidation Report:\n%s", lastOutput, lastReport.Summary),
		Diff:     e.server.GetDiff(targetPath),
	}

	reporter := &Reporter{}
	res.Report = reporter.GenerateReport(res, workspace)

	return res, nil
}

func (e *Executor) assessRisk(mt *MutationTask) RiskLevel {
	switch mt.Mutation.Type {
	case "rename", "cleanup":
		return RiskLow
	case "replace":
		return RiskMedium
	case "inject":
		return RiskHigh
	default:
		return RiskMedium
	}
}
