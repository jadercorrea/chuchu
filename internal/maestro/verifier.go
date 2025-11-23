package maestro

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"chuchu/internal/langdetect"
)

type VerificationResult struct {
	Success bool
	Output  string
	Error   error
}

type Verifier interface {
	Verify(ctx context.Context) (*VerificationResult, error)
}

type TestVerifier struct {
	Dir      string
	Language string
}

func NewTestVerifier(dir string) *TestVerifier {
	lang := detectLanguage(dir)
	return &TestVerifier{Dir: dir, Language: lang}
}

func (v *TestVerifier) Verify(ctx context.Context) (*VerificationResult, error) {
	var cmd *exec.Cmd

	switch v.Language {
	case "go":
		cmd = exec.CommandContext(ctx, "go", "test", "./...")
	case "javascript", "typescript":
		if fileExists(filepath.Join(v.Dir, "package.json")) {
			cmd = exec.CommandContext(ctx, "npm", "test")
		} else {
			return &VerificationResult{Success: true}, nil
		}
	case "python":
		if fileExists(filepath.Join(v.Dir, "pytest.ini")) || fileExists(filepath.Join(v.Dir, "setup.py")) {
			cmd = exec.CommandContext(ctx, "pytest")
		} else {
			return &VerificationResult{Success: true}, nil
		}
	case "elixir":
		cmd = exec.CommandContext(ctx, "mix", "test")
	case "ruby":
		if fileExists(filepath.Join(v.Dir, "Gemfile")) {
			cmd = exec.CommandContext(ctx, "bundle", "exec", "rspec")
		} else {
			cmd = exec.CommandContext(ctx, "rspec")
		}
	default:
		return &VerificationResult{Success: true}, nil
	}

	cmd.Dir = v.Dir
	output, err := cmd.CombinedOutput()
	outStr := string(output)

	if err != nil {
		return &VerificationResult{
			Success: false,
			Output:  outStr,
			Error:   fmt.Errorf("tests failed: %w", err),
		}, nil
	}

	return &VerificationResult{
		Success: true,
		Output:  outStr,
		Error:   nil,
	}, nil
}

type BuildVerifier struct {
	Dir      string
	Language string
}

func NewBuildVerifier(dir string) *BuildVerifier {
	lang := detectLanguage(dir)
	return &BuildVerifier{Dir: dir, Language: lang}
}

func (v *BuildVerifier) Verify(ctx context.Context) (*VerificationResult, error) {
	var cmd *exec.Cmd

	switch v.Language {
	case "go":
		cmd = exec.CommandContext(ctx, "go", "build", "./...")
	case "javascript", "typescript":
		if fileExists(filepath.Join(v.Dir, "package.json")) {
			cmd = exec.CommandContext(ctx, "npm", "run", "build")
		} else {
			return &VerificationResult{Success: true}, nil
		}
	case "python":
		cmd = exec.CommandContext(ctx, "python", "-m", "py_compile")
	case "elixir":
		cmd = exec.CommandContext(ctx, "mix", "compile")
	case "ruby":
		return &VerificationResult{Success: true}, nil
	default:
		return &VerificationResult{Success: true}, nil
	}

	cmd.Dir = v.Dir
	output, err := cmd.CombinedOutput()
	outStr := string(output)

	if err != nil {
		return &VerificationResult{
			Success: false,
			Output:  outStr,
			Error:   fmt.Errorf("build failed: %w", err),
		}, nil
	}

	return &VerificationResult{
		Success: true,
		Output:  outStr,
		Error:   nil,
	}, nil
}

func detectLanguage(dir string) string {
	lang := langdetect.DetectLanguage(dir)
	switch lang {
	case langdetect.Go:
		return "go"
	case langdetect.TypeScript:
		return "typescript"
	case langdetect.Python:
		return "python"
	case langdetect.Elixir:
		return "elixir"
	case langdetect.Ruby:
		return "ruby"
	default:
		return "unknown"
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
