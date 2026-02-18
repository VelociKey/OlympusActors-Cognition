package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"Olympus2/90000-Enablement-Labs/P0000-pkg/000-mesh"
	"Olympus2/90000-Enablement-Labs/P0000-pkg/000-whisper"
)

// ArchitectAgent: The Structural Pillar (Olympus2 Standard)
const WorkspaceRoot = "c:\\aAntigravitySpace\\Olympus2"

var sc *whisper.WhisperLog

func main() {
	logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	slog.SetDefault(slog.New(logHandler))

	meshHubURL := os.Getenv("MESH_HUB_URL")
	if meshHubURL == "" {
		meshHubURL = "http://localhost:8090"
	}

	sc = whisper.New("Architect", "architect.lpsv")

	mux := http.NewServeMux()

	mux.HandleFunc("/pulse", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "ArchitectAgent: ACTIVE. Role: Guardian of Structure. Status: Healthy.")
		sc.Log("Pulse", "Success", "heartbeat", "Agent is healthy", 0)
	})

	mux.HandleFunc("/audit", func(w http.ResponseWriter, r *http.Request) {
		slog.Info("🏗️ Architect: Initiating Sovereign Audit")
		report := performAudit()
		sc.Log("PerformAudit", "Success", WorkspaceRoot, report, 0)
		fmt.Fprintf(w, "Architect: AUDIT_COMPLETE.\n")
		fmt.Fprintf(w, "Report: %s", report)
	})

	srv := &http.Server{
		Addr:         ":8085",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		// Self-Register
		time.Sleep(1 * time.Second)
		mesh.RegisterWithMesh(context.Background(), meshHubURL, "Architect", 8085, "guardian", []string{"audit", "structural-verification"})

		slog.Info("ArchitectAgent starting", "port", "8085")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed", "error", err)
			os.Exit(1)
		}
	}()

	<-stop
	slog.Info("Shutting down gracefully...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(shutdownCtx)
	sc.Close()
	slog.Info("Server stopped")
}

func performAudit() string {
	domains := []string{"Olympus2/00000", "Olympus2/10000", "Olympus2/20000", "Olympus2/30000", "Olympus2/40000", "Olympus2/50000", "Olympus2/60000", "Olympus2/70000", "Olympus2/80000", "Olympus2/90000"}
	missing := 0
	for _, domain := range domains {
		path := filepath.Join(WorkspaceRoot, domain+"-*")
		matches, _ := filepath.Glob(path)
		if len(matches) == 0 {
			if _, err := os.Stat(filepath.Join(WorkspaceRoot, domain)); os.IsNotExist(err) {
				missing++
			}
		}
	}
	hasGoWork := false
	if _, err := os.Stat(filepath.Join(WorkspaceRoot, "go.work")); err == nil {
		hasGoWork = true
	}
	return fmt.Sprintf("Domains Identified: %d/10. go.work Status: %t. State: SECURE.", 10-missing, hasGoWork)
}
