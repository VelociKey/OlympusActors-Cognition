package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// BlueprintStats tracks the performance of a specific mutation blueprint.
type BlueprintStats struct {
	Name           string
	Executions     int
	Successes      int
	Failures       int
	AvgEfficacy    float32
	LastExecution  time.Time
	FailureReasons []string
}

// LearningEngine (LE): Optimized for jeBNF Persistence
type LearningEngine struct {
	StatsPath string
	Stats     map[string]*BlueprintStats
	mu        sync.RWMutex
}

func NewLearningEngine(path string) *LearningEngine {
	// Transition to .jebnf extension
	jebnfPath := strings.TrimSuffix(path, ".json") + ".jebnf"
	le := &LearningEngine{
		StatsPath: jebnfPath,
		Stats:     make(map[string]*BlueprintStats),
	}
	le.load()
	return le
}

func (le *LearningEngine) RecordResult(blueprintName string, result *PipelineResult) {
	le.mu.Lock()
	defer le.mu.Unlock()

	stats, exists := le.Stats[blueprintName]
	if !exists {
		stats = &BlueprintStats{Name: blueprintName}
		le.Stats[blueprintName] = stats
	}

	stats.Executions++
	if result.Status == "SUCCESS" {
		stats.Successes++
	} else {
		stats.Failures++
		stats.FailureReasons = append(stats.FailureReasons, "Failure detected")
	}

	stats.AvgEfficacy = (stats.AvgEfficacy*float32(stats.Executions-1) + result.Efficacy) / float32(stats.Executions)
	stats.LastExecution = time.Now()

	slog.Info("🧠 LearningEngine: Updated stats", "blueprint", blueprintName)
	le.save()
}

func (le *LearningEngine) load() {
	le.mu.Lock()
	defer le.mu.Unlock()

	f, err := os.Open(le.StatsPath)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var current *BlueprintStats
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") { continue }

		if strings.HasPrefix(line, "Blueprint ") {
			name := strings.TrimSuffix(strings.TrimPrefix(line, "Blueprint "), " {")
			current = &BlueprintStats{Name: name}
			le.Stats[name] = current
		} else if current != nil && strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			k := strings.TrimSpace(parts[0])
			v := strings.Trim(strings.TrimSpace(parts[1]), "\"")

			switch k {
			case "executions": current.Executions, _ = strconv.Atoi(v)
			case "successes":  current.Successes, _ = strconv.Atoi(v)
			case "failures":   current.Failures, _ = strconv.Atoi(v)
			case "efficacy":   eff, _ := strconv.ParseFloat(v, 32); current.AvgEfficacy = float32(eff)
			case "last_run":   current.LastExecution, _ = time.Parse(time.RFC3339, v)
			}
		}
	}
}

func (le *LearningEngine) save() {
	var sb strings.Builder
	sb.WriteString("# CoderAgent Learning Stats (jeBNF)\n\n")

	for name, s := range le.Stats {
		sb.WriteString(fmt.Sprintf("Blueprint %s {\n", name))
		sb.WriteString(fmt.Sprintf("  executions = %d\n", s.Executions))
		sb.WriteString(fmt.Sprintf("  successes = %d\n", s.Successes))
		sb.WriteString(fmt.Sprintf("  failures = %d\n", s.Failures))
		sb.WriteString(fmt.Sprintf("  efficacy = %.2f\n", s.AvgEfficacy))
		sb.WriteString(fmt.Sprintf("  last_run = \"%s\"\n", s.LastExecution.Format(time.RFC3339)))
		sb.WriteString("}\n\n")
	}

	os.MkdirAll(filepath.Dir(le.StatsPath), 0755)
	os.WriteFile(le.StatsPath, []byte(sb.String()), 0644)
}

func (le *LearningEngine) GenerateMaturityReport() string {
	le.mu.RLock()
	defer le.mu.RUnlock()
	report := "# 📈 CoderAgent Maturity & Learning Report\n\n"
	for name, s := range le.Stats {
		rate := float32(s.Successes) / float32(s.Executions) * 100
		report += fmt.Sprintf("### %s\n- Success Rate: %.1f%%\n- Avg Efficacy: %.1f/5.0\n- Last Run: %s\n\n",
			name, rate, s.AvgEfficacy, s.LastExecution.Format("2006-01-02"))
	}
	return report
}
