package maestro

import (
	"context"
	"os/exec"
	"path/filepath"
)

type LintVerifier struct {
	Dir      string
	Language string
}

func NewLintVerifier(dir string) *LintVerifier {
	lang := detectLanguage(dir)
	return &LintVerifier{Dir: dir, Language: lang}
}

func (v *LintVerifier) Verify(ctx context.Context) (*VerificationResult, error) {
	var cmd *exec.Cmd

	switch v.Language {
	case "go":
		if commandExists("golangci-lint") {
			cmd = exec.CommandContext(ctx, "golangci-lint", "run", "./...")
		} else {
			cmd = exec.CommandContext(ctx, "go", "vet", "./...")
		}
	case "javascript", "typescript":
		if fileExists(filepath.Join(v.Dir, ".eslintrc.json")) || fileExists(filepath.Join(v.Dir, ".eslintrc.js")) {
			if commandExists("eslint") {
				cmd = exec.CommandContext(ctx, "npm", "run", "lint")
			}
		}
	case "python":
		if commandExists("ruff") {
			cmd = exec.CommandContext(ctx, "ruff", "check", ".")
		} else if commandExists("flake8") {
			cmd = exec.CommandContext(ctx, "flake8", ".")
		}
	case "elixir":
		cmd = exec.CommandContext(ctx, "mix", "format", "--check-formatted")
	case "ruby":
		if commandExists("rubocop") {
			cmd = exec.CommandContext(ctx, "rubocop")
		}
	default:
		return &VerificationResult{Success: true}, nil
	}

	if cmd == nil {
		return &VerificationResult{Success: true}, nil
	}

	cmd.Dir = v.Dir
	output, err := cmd.CombinedOutput()
	outStr := string(output)

	if err != nil {
		return &VerificationResult{
			Success: false,
			Output:  outStr,
			Error:   err,
		}, nil
	}

	return &VerificationResult{
		Success: true,
		Output:  outStr,
	}, nil
}

func commandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
