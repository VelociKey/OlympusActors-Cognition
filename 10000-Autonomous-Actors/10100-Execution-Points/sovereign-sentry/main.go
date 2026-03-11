package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Sovereign Sentry - Evolutionary Optimizer
// Job: Evaluate script failures and drive toolchain liquidation.

func main() {
	// 1. Setup Structured Telemetry (LPSV-compliant slog)
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	slog.SetDefault(logger)

	slog.Info("WORKING [Sentry] Initializing Daily Evolutionary Review")

	// 2. Identify fail sensors
	errorLog := "00SDLC/Olympus2/C0990-Ephemeral-Scratch/script-errors.jebnf"
	liquidationReqs := "conductor/liquidation-requests.jebnf"

	if _, err := os.Stat(errorLog); os.IsNotExist(err) {
		slog.Info("COMPLETE [Sentry] No script errors found. Toolchain healthy.")
		os.Exit(0)
	}

	// 3. Evaluate Failures
	slog.Info("WORKING [Sentry] Clustering failure patterns")
	clusters := evaluateFailures(errorLog)

	if len(clusters) == 0 {
		slog.Info("COMPLETE [Sentry] Review finished. No actionable liquidation required.")
		os.Exit(0)
	}

	// 4. Generate Liquidation Requests
	slog.Info("WORKING [Sentry] Generating liquidation requests", "count", len(clusters))
	err := generateRequests(liquidationReqs, clusters)
	if err != nil {
		slog.Error("FAILED [Sentry] Failed to generate liquidation requests", "error", err)
		os.Exit(1)
	}

	slog.Info("COMPLETE [Sentry] Daily review finalized. Proposals emitted to Conductor.")
}

type FailureCluster struct {
	Signature string
	Count     int
	Examples  []string
}

func evaluateFailures(path string) map[string]*FailureCluster {
	clusters := make(map[string]*FailureCluster)
	file, err := os.Open(path)
	if err != nil {
		return clusters
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Simple clustering: look for PowerShell syntax errors or command not found
		sig := "unknown-failure"
		if strings.Contains(line, "ParserError") || strings.Contains(line, "&&") {
			sig = "shell-syntax-incompatibility"
		} else if strings.Contains(line, "not recognized as the name of a cmdlet") {
			sig = "missing-native-binary"
		}

		if _, ok := clusters[sig]; !ok {
			clusters[sig] = &FailureCluster{Signature: sig, Examples: []string{}}
		}
		clusters[sig].Count++
		if len(clusters[sig].Examples) < 3 {
			clusters[sig].Examples = append(clusters[sig].Examples, line)
		}
	}
	return clusters
}

func generateRequests(path string, clusters map[string]*FailureCluster) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	timestamp := time.Now().Format(time.RFC3339)
	for sig, cluster := range clusters {
		entry := fmt.Sprintf("\nLiquidationRequest {\n  Timestamp = \"%s\";\n  Pattern = \"%s\";\n  Occurrences = %d;\n  Priority = \"HIGH\";\n  Initiator = \"AI\";\n  Status = \"PROPOSED\";\n}\n",
			timestamp, sig, cluster.Count)
		_, err := f.WriteString(entry)
		if err != nil {
			return err
		}
	}
	return nil
}
