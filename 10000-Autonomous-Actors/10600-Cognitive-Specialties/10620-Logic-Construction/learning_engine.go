package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type BlueprintStats struct {
	Executions    int       `json:"executions"`
	Successes     int       `json:"successes"`
	Failures      int       `json:"failures"`
	AvgEfficacy   float32   `json:"avg_efficacy"`
	LastExecution time.Time `json:"last_execution"`
}

type LearningEngine struct {
	Stats     map[string]*BlueprintStats
	StatsPath string
	mu        sync.RWMutex
}

func NewLearningEngine(path string) *LearningEngine {
	le := &LearningEngine{
		Stats:     make(map[string]*BlueprintStats),
		StatsPath: path,
	}
	le.load()
	return le
}

func (le *LearningEngine) load() {
	data, err := os.ReadFile(le.StatsPath)
	if err != nil {
		return
	}

	// This is a simple parser for the jeBNF-style stats file
	content := string(data)
	lines := strings.Split(content, "\n")
	var currentName string
	var currentStats *BlueprintStats

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Blueprint ") {
			currentName = strings.TrimPrefix(strings.TrimSuffix(line, " {"), "Blueprint ")
			currentStats = &BlueprintStats{}
			le.Stats[currentName] = currentStats
		} else if currentStats != nil {
			if strings.HasPrefix(line, "executions = ") {
				fmt.Sscanf(line, "executions = %d", &currentStats.Executions)
			} else if strings.HasPrefix(line, "successes = ") {
				fmt.Sscanf(line, "successes = %d", &currentStats.Successes)
			} else if strings.HasPrefix(line, "failures = ") {
				fmt.Sscanf(line, "failures = %d", &currentStats.Failures)
			} else if strings.HasPrefix(line, "efficacy = ") {
				fmt.Sscanf(line, "efficacy = %f", &currentStats.AvgEfficacy)
			} else if strings.HasPrefix(line, "last_run = ") {
				var ts string
				fmt.Sscanf(line, "last_run = \"%s\"", &ts)
				ts = strings.Trim(ts, "\"")
				currentStats.LastExecution, _ = time.Parse(time.RFC3339, ts)
			}
		}
	}
}

func (le *LearningEngine) RecordResult(blueprint string, success bool, efficacy float32) {
	le.mu.Lock()
	defer le.mu.Unlock()

	s, ok := le.Stats[blueprint]
	if !ok {
		s = &BlueprintStats{}
		le.Stats[blueprint] = s
	}

	s.Executions++
	if success {
		s.Successes++
	} else {
		s.Failures++
	}

	// Rolling average for efficacy
	if s.Executions == 1 {
		s.AvgEfficacy = efficacy
	} else {
		s.AvgEfficacy = (s.AvgEfficacy*float32(s.Executions-1) + efficacy) / float32(s.Executions)
	}
	s.LastExecution = time.Now()

	le.save()
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
		report += fmt.Sprintf("### %s\n- Success Rate: %.1f%%\n- Avg Efficacy: %.1f/5.0\n- Last Run: %s\n\n", name, rate, s.AvgEfficacy, s.LastExecution.Format("2006-01-02"))
	}
	return report
}
