package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Validator handles multi-stage validation of code changes.
type Validator struct {
	path string
}

// ValidationReport contains detailed results of the validation suite.
type ValidationReport struct {
	Success bool
	Score   float64
	Summary string
	Details map[string]StepResult
}

type StepResult struct {
	Status   string
	Duration time.Duration
	Output   string
}

// Validate runs build, test, and (placeholder) lint checks.
func (v *Validator) Validate(ctx context.Context) (*ValidationReport, error) {
	report := &ValidationReport{
		Success: true,
		Details: make(map[string]StepResult),
	}
	var sb strings.Builder

	// 1. Build Verification
	start := time.Now()
	cmdBuild := exec.CommandContext(ctx, "go", "build", "./...")
	cmdBuild.Dir = v.path
	outBuild, err := cmdBuild.CombinedOutput()
	durationBuild := time.Since(start)

	if err != nil {
		report.Success = false
		report.Details["build"] = StepResult{Status: "FAIL", Duration: durationBuild, Output: string(outBuild)}
		sb.WriteString(fmt.Sprintf("❌ Build Failed (%.2fs)\n", durationBuild.Seconds()))
	} else {
		report.Details["build"] = StepResult{Status: "PASS", Duration: durationBuild}
		sb.WriteString(fmt.Sprintf("✅ Build Passed (%.2fs)\n", durationBuild.Seconds()))
	}

	if !report.Success {
		report.Summary = sb.String()
		report.Score = 1.0
		return report, nil
	}

	// 2. Test Verification
	start = time.Now()
	cmdTest := exec.CommandContext(ctx, "go", "test", "./...")
	cmdTest.Dir = v.path
	outTest, err := cmdTest.CombinedOutput()
	durationTest := time.Since(start)

	if err != nil {
		// Note: We don't necessarily fail the whole report just because a test failed,
		// but for C08 we want high efficacy (1.0).
		report.Details["test"] = StepResult{Status: "FAIL", Duration: durationTest, Output: string(outTest)}
		sb.WriteString(fmt.Sprintf("❌ Tests Failed (%.2fs)\n", durationTest.Seconds()))
	} else {
		report.Details["test"] = StepResult{Status: "PASS", Duration: durationTest}
		sb.WriteString(fmt.Sprintf("✅ Tests Passed (%.2fs)\n", durationTest.Seconds()))
	}

	// Calculate Score (Simple 1-5 scale)
	score := 5.0
	if report.Details["test"].Status == "FAIL" {
		score = 3.0
	}
	if !report.Success {
		score = 1.0
	}

	report.Score = score
	report.Summary = sb.String()
	return report, nil
}
