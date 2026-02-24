package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"connectrpc.com/connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	olympusv1 "Olympus2/gen/v1/olympus"
	olympusv1connect "Olympus2/gen/v1/olympus/olympusv1connect"
	"Olympus2/90000-Enablement-Labs/P0000-pkg/000-mesh"
	"Olympus2/90000-Enablement-Labs/P0000-pkg/000-whisper"
)

var SandboxRoot = getEnv("WORKSPACE_ROOT", "c:\\aAntigravitySpace")

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}

type CoderServer struct {
	olympusv1connect.UnimplementedCoderServiceHandler
	olympusv1connect.UnimplementedMissionServiceHandler
	memoryClient    olympusv1connect.MemoryServiceClient
	knowledgeClient olympusv1connect.KnowledgeServiceClient
	inferenceClient olympusv1connect.InferenceServiceClient
	auditClient     olympusv1connect.AuditServiceClient
	engine          *MutationEngine
	executor        *Executor
	sc              *whisper.WhisperLog
	agentToken      string
}

func (s *CoderServer) ExecuteTask(ctx context.Context, req *connect.Request[olympusv1.TaskRequest]) (*connect.Response[olympusv1.TaskResponse], error) {
	meta := mesh.FromContext(ctx)
	if !meta.HasCapability("mutation") && !meta.HasCapability("*") {
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("missing mutation capability"))
	}
	targetPath, err := validateSandbox(req.Msg.Workspace)
	if err != nil {
		return nil, connect.NewError(connect.CodePermissionDenied, err)
	}
	slog.Info("🤖 Coder: Task Active", "mission", meta.MissionID, "agent_id", meta.AgentID)

	if err := s.checkpoint(targetPath); err != nil {
		slog.Error("Failed to checkpoint workspace", "error", err)
	}

	output, nlErr := s.RunGemini(ctx, targetPath, req.Msg.Task, req.Msg.PlanOnly)
	status := "SUCCESS"
	if nlErr != nil {
		status = "FAILURE"
	}
	if status == "SUCCESS" && !req.Msg.PlanOnly {
		if err := s.commit(targetPath); err != nil {
			slog.Error("Failed to commit mutation", "error", err)
		}
		_, err := s.memoryClient.LogEvent(ctx, connect.NewRequest(&olympusv1.EventRequest{
			Agent: "Coder", Action: "mutation", Target: req.Msg.Workspace, Status: "Success",
			TraceId: meta.TraceID, MissionId: meta.MissionID, AgentId: s.agentToken,
		}))
		if err != nil {
			slog.Error("Failed to log mutation event", "error", err)
		}
	}
	return connect.NewResponse(&olympusv1.TaskResponse{Status: status, Output: output}), nil
}

func (s *CoderServer) ExecuteMission(ctx context.Context, req *connect.Request[olympusv1.MissionRequest]) (*connect.Response[olympusv1.MissionResponse], error) {
	meta := mesh.FromContext(ctx)
	manifestPath := filepath.Clean(req.Msg.ManifestPath)
	workspace := req.Msg.Workspace
	slog.Info("🚀 Coder: Executing Mission", "manifest", manifestPath, "mission", meta.MissionID)

	targetPath, err := validateSandbox(workspace)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	fullPath := filepath.Join(targetPath, manifestPath)
	rel, err := filepath.Rel(targetPath, fullPath)
	if err != nil || strings.HasPrefix(rel, "..") {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid manifest path"))
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	prompt := fmt.Sprintf("PARSE THIS jeBNF MISSION INTO A JSON LIST OF STEPS:\n%s", string(data))
	_, err = s.inferenceClient.Reason(ctx, connect.NewRequest(&olympusv1.ReasonRequest{Prompt: prompt, Context: map[string]string{"blueprint": "orchestrator"}}))
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	workspaces := []string{"Olympus2", "OlympusFabric", "OlympusAtelier", "OlympusGrammar", "OlympusMuse"}
	var results []*olympusv1.MissionStep
	for _, ws := range workspaces {
		slog.Info("🛡️ Mission: Auditing scope", "ws", ws)
		auditRes, err := s.auditClient.Assess(ctx, connect.NewRequest(&olympusv1.AuditRequest{Workspace: ws}))
		status := "SUCCESS"
		if err != nil || auditRes == nil {
			status = "FAILURE"
			slog.Error("Mission audit step failed", "workspace", ws, "error", err)
		}
		results = append(results, &olympusv1.MissionStep{Id: ws, TargetWorkspace: ws, Status: status, Output: "Maturity Scored"})
	}
	return connect.NewResponse(&olympusv1.MissionResponse{MissionId: meta.MissionID, Status: "COMPLETED", Results: results}), nil
}

func (s *CoderServer) RunGemini(ctx context.Context, path, task string, planOnly bool) (string, error) {
	res, err := s.inferenceClient.Reason(ctx, connect.NewRequest(&olympusv1.ReasonRequest{Prompt: task, Context: map[string]string{"blueprint": "coder", "mode": "system2"}}))
	if err != nil {
		return "", err
	}
	return res.Msg.Output, nil
}

func (s *CoderServer) checkpoint(path string) error {
	cmd := exec.Command("git", "stash")
	cmd.Dir = path
	return cmd.Run()
}

func (s *CoderServer) commit(path string) error {
	cmdAdd := exec.Command("git", "add", ".")
	cmdAdd.Dir = path
	if err := cmdAdd.Run(); err != nil {
		return err
	}

	cmdCommit := exec.Command("git", "commit", "-m", "chore: mutation")
	cmdCommit.Dir = path
	return cmdCommit.Run()
}

func (s *CoderServer) GetDiff(path string) string {
	cmd := exec.Command("git", "diff", "HEAD~1")
	cmd.Dir = path
	out, err := cmd.Output()
	if err != nil {
		slog.Error("Failed to get git diff", "error", err)
		return ""
	}
	return string(out)
}

func (s *CoderServer) TaskJules(ctx context.Context, req *connect.Request[olympusv1.TaskRequest]) (*connect.Response[olympusv1.TaskResponse], error) {
	workspace := req.Msg.Workspace
	slog.Info("🤖 Coder: Tasking Jules for tests", "workspace", workspace)

	targetPath, err := validateSandbox(workspace)
	if err != nil {
		return nil, connect.NewError(connect.CodePermissionDenied, err)
	}

	go func() {
		err := taskJulesForTests(targetPath)
		if err != nil {
			slog.Error("Jules task failed", "workspace", workspace, "error", err)
		}
	}()

	return connect.NewResponse(&olympusv1.TaskResponse{
		Status: "ACCEPTED",
		Output: fmt.Sprintf("Jules background campaign started for %s", workspace),
	}), nil
}

func taskJulesForTests(path string) error {
	// Pattern: jules /generate:tests --target="<Path>" --type="unit,integration"
	cmd := exec.Command("jules", "/generate:tests", "--target="+path, "--type=unit,integration")
	return cmd.Run()
}

func main() {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	slog.SetDefault(slog.New(handler))
	meshHubURL := getEnv("MESH_HUB_URL", "http://localhost:8090")
	guardianURL := getEnv("GUARDIAN_URL", "http://localhost:8082")
	interceptors := connect.WithInterceptors(mesh.NewInterceptor(guardianURL))
	token, _ := mesh.Handshake(context.Background(), guardianURL, "Coder", "HW-WIN-01", []string{"mission_control", "mutation"})
	server := &CoderServer{
		agentToken:      token,
		memoryClient:    olympusv1connect.NewMemoryServiceClient(http.DefaultClient, getEnv("MEMORY_URL", "http://localhost:8084"), interceptors),
		knowledgeClient: olympusv1connect.NewKnowledgeServiceClient(http.DefaultClient, getEnv("CARTOGRAPHER_URL", "http://localhost:8095"), interceptors),
		inferenceClient: olympusv1connect.NewInferenceServiceClient(http.DefaultClient, getEnv("INFERENCE_URL", "http://localhost:8087"), interceptors),
		auditClient:     olympusv1connect.NewAuditServiceClient(http.DefaultClient, getEnv("AUDIT_URL", "http://localhost:8086"), interceptors),
		sc:              whisper.New("Coder", "coder.lpsv"),
	}
	server.engine = &MutationEngine{server: server}
	server.executor = &Executor{server: server}
	mux := http.NewServeMux()

	// Health Check / Pulse
	mux.HandleFunc("/pulse", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"HEALTHY", "workspace":"OlympusActors-Cognition", "time":"%s"}`, time.Now().Format(time.RFC3339))
	})

	pathCoder, handlerCoder := olympusv1connect.NewCoderServiceHandler(server, interceptors)
	mux.Handle(pathCoder, handlerCoder)

	pathMission, handlerMission := olympusv1connect.NewMissionServiceHandler(server, interceptors)
	mux.Handle(pathMission, handlerMission)

	srv := &http.Server{
		Addr:              ":8083",
		Handler:           h2c.NewHandler(mux, &http2.Server{}),
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		time.Sleep(1 * time.Second)
		if err := mesh.RegisterWithMesh(context.Background(), meshHubURL, "Coder", 8083, "mutation", []string{"mission-orchestration"}); err != nil {
			slog.Error("Failed to register with mesh", "error", err)
		}
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed", "error", err)
		}
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	if err := srv.Shutdown(context.Background()); err != nil {
		slog.Error("Server shutdown error", "error", err)
	}
}

func validateSandbox(workspace string) (string, error) {
	target := filepath.Join(SandboxRoot, filepath.Clean(workspace))
	rel, err := filepath.Rel(SandboxRoot, target)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("invalid workspace path: traversal detected")
	}
	if _, err := os.Stat(target); os.IsNotExist(err) {
		return "", fmt.Errorf("workspace does not exist")
	}
	return target, nil
}

