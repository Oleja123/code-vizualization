package analyzer

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type Issue struct {
	Severity string `json:"severity"`
	ID       string `json:"id"`
	Line     int    `json:"line"`
	Message  string `json:"message"`
}

type Analyzer struct {
	path         string
	std          string
	enable       string
	inconclusive bool
	maxIssues    int
}

func New(path, std, enable string, inconclusive bool, maxIssues int) *Analyzer {
	return &Analyzer{
		path:         path,
		std:          std,
		enable:       enable,
		inconclusive: inconclusive,
		maxIssues:    maxIssues,
	}
}

func (a *Analyzer) Analyze(ctx context.Context, code string) ([]Issue, string, error) {
	tmpDir, err := os.MkdirTemp("", "cppcheck-analyzer-*")
	if err != nil {
		return nil, "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "input.c")
	if err := os.WriteFile(filePath, []byte(code), 0o600); err != nil {
		return nil, "", fmt.Errorf("failed to write source file: %w", err)
	}

	args := []string{
		"--language=c",
		"--std=" + a.std,
		"--template={severity}|{id}|{line}|{message}",
		"--enable=" + a.enable,
		"--quiet",
	}

	if a.inconclusive {
		args = append(args, "--inconclusive")
	}

	args = append(args, filePath)

	cmd := exec.CommandContext(ctx, a.path, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	execErr := cmd.Run()
	rawOutput := strings.TrimSpace(strings.TrimSpace(stdout.String() + "\n" + stderr.String()))
	issues := parseCppcheckOutput(rawOutput, a.maxIssues)

	if execErr != nil {
		if len(issues) > 0 {
			return issues, rawOutput, nil
		}
		return nil, rawOutput, fmt.Errorf("failed to execute cppcheck: %w", execErr)
	}

	return issues, rawOutput, nil
}

func parseCppcheckOutput(raw string, maxIssues int) []Issue {
	if strings.TrimSpace(raw) == "" {
		return nil
	}

	lines := strings.Split(raw, "\n")
	issues := make([]Issue, 0, len(lines))

	for _, line := range lines {
		if maxIssues > 0 && len(issues) >= maxIssues {
			break
		}

		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		parts := strings.SplitN(trimmed, "|", 4)
		if len(parts) != 4 {
			// Ignore non-template output lines (tool logs, progress messages).
			continue
		}

		lineNumber, err := strconv.Atoi(strings.TrimSpace(parts[2]))
		if err != nil || lineNumber <= 0 {
			// Keep only issues that can be mapped to concrete source code lines.
			continue
		}

		severity := strings.TrimSpace(parts[0])
		id := strings.TrimSpace(parts[1])
		message := strings.TrimSpace(parts[3])
		if severity == "" || id == "" || message == "" {
			continue
		}

		issues = append(issues, Issue{
			Severity: severity,
			ID:       id,
			Line:     lineNumber,
			Message:  message,
		})
	}

	sort.Slice(issues, func(i, j int) bool {
		return issues[i].Line < issues[j].Line
	})

	return issues
}
