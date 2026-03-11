package main

import (
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	whisper "olympus.fleet/00SDLC/Olympus2/90000-Enablement-Labs/90200-Logic-Libraries/220-Whisper"
)

// Sovereign Sentry - Evolutionary Optimizer
// Job: Evaluate script failures and drive toolchain liquidation.

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	slog.SetDefault(logger)

	slog.Info("WORKING [Sentry] Starting Daily Evolutionary Review")

	// Invoke GemAid for the heavy lifting
	gemaidPath := "00SDLC/OlympusForge/82000-Toolchain-Fleet/bin/gemaid.exe"
	cmd := exec.Command(gemaidPath, "sentry", "evaluate")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		slog.Error("FAILED [Sentry] GemAid sentry evaluation failed", "error", err)
		// Continue to monitoring despite evaluation failure
	}

	slog.Info("COMPLETE [Sentry] Daily review finalized. Starting Real-time Watchdog...")

	// Locate MeshHub Log
	logDir := os.Getenv("WHISPER_LOG_DIR")
	if logDir == "" {
		// Fallback to absolute workspace standard path if env is missing
		logDir = "C:/aAntigravitySpace/olympus.fleet/00SDLC/Olympus2/C0500-Agent-Intelligence-Outputs/LPSV"
	}
	meshLog := filepath.Join(logDir, "meshhub.lpsv")

	// Start Listening
	events := whisper.NewListener(meshLog)
	for event := range events {
		if strings.Contains(event, "OFFLINE") {
			slog.Warn("🚨 AGENT OFFLINE DETECTED", "event", event)
			// Future: Trigger remediation or alert via EventBus
		}
	}
}
