package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"olympus.fleet/00SDLC/OlympusFabric/40000-Communication-Contracts/40400-Protocol-Synthetics/connect-rpc/v1/agent/agentv1connect"

	mesh "olympus.fleet/00SDLC/Olympus2/90000-Enablement-Labs/90200-Logic-Libraries/150-Mesh"
)

func main() {
	logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	slog.SetDefault(slog.New(logHandler))

	meshHubURL := os.Getenv("MESH_HUB_URL")
	if meshHubURL == "" {
		meshHubURL = "http://localhost:8090"
	}

	httpClient := http.DefaultClient
	planner := &MissionPlanner{
		memory:       agentv1connect.NewMemoryServiceClient(httpClient, getEnv("MEMORY_URL", "http://localhost:8082")),
		intelligence: agentv1connect.NewInferenceServiceClient(httpClient, getEnv("INFERENCE_URL", "http://localhost:8081")),
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/propose", func(w http.ResponseWriter, r *http.Request) {
		objective := r.URL.Query().Get("objective")
		if objective == "" {
			objective = "Autonomous Strategic Alignment"
		}

		slog.Info("Agent: Initiating Mission Proposal", "objective", objective)
		resp, err := planner.ProposeNextMission(context.Background(), objective)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "Proposed Mission:\n%s", resp.Summary)
	})

	srv := &http.Server{
		Addr:         ":8086",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		// Self-Register
		time.Sleep(1 * time.Second)
		mesh.RegisterWithMesh(context.Background(), meshHubURL, "MissionPlanner", 8086, "cognitive", []string{"planning", "mission-generation"})

		slog.Info("MissionPlanner starting", "port", "8086")
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
	slog.Info("Server stopped")
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
