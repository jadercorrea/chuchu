package agents

import (
	"context"
	"fmt"
	"os"
	"strings"

	"chuchu/internal/llm"
	"chuchu/internal/tools"
)

type ReviewerAgent struct {
	provider llm.Provider
	cwd      string
	model    string
}

type ReviewResult struct {
	Success     bool
	Issues      []string
	Suggestions string
}

type ReviewerConfig struct {
	OnValidationFail func(issues []string) (shouldRetry bool, newBackend string, newModel string)
}

func NewReviewer(provider llm.Provider, cwd string, model string) *ReviewerAgent {
	return &ReviewerAgent{
		provider: provider,
		cwd:      cwd,
		model:    model,
	}
}

const reviewerPrompt = `You are a code reviewer. Your job is to verify if changes meet the success criteria.

WORKFLOW:
1. Read the files that were modified
2. Run commands to verify (build, test, lint, etc)
3. Compare against the success criteria
4. Report pass/fail with specific issues

CRITICAL RULES:
- ALWAYS run build/compile commands to verify code compiles
- For Go: run 'go build' to check compilation
- For TypeScript/Node: run 'npm run build' or 'tsc'
- For Python: check syntax with 'python -m py_compile file.py'
- Be specific about what's wrong
- If something is missing, say exactly what
- If criteria is met, say "SUCCESS"
- Focus on the actual requirements, not style

EXAMPLE 1 - Validation SUCCESS:
Success Criteria:
  - Tests pass: go test ./auth/...
  - File auth/handler.go contains Login function
  - Lints clean: golangci-lint run

Validation:
  1. Read auth/handler.go → contains func Login()
  2. Would run: go test ./auth/... → (assume passes)
  3. Would run: golangci-lint run → (assume clean)

Result: SUCCESS

EXAMPLE 2 - Validation FAIL with specific issues:
Success Criteria:
  - Tests pass: make test
  - File middleware/jwt.go exports VerifyToken function
  - No hardcoded secrets

Validation:
  1. Read middleware/jwt.go
  
Issues found:
  - FAIL: VerifyToken is not exported (lowercase verifyToken)
  - FAIL: JWT secret is hardcoded on line 15: "hardcoded-secret-123"
  - Tests: Would need to run make test to verify

Result:
FAIL

Issues:
- VerifyToken function must be exported (capitalize: VerifyToken)
- Remove hardcoded secret on line 15, use environment variable
- Run make test to verify tests pass

EXAMPLE 3 - Clear issue reporting:
BAD:
  "Something is wrong with the file"
  
GOOD:
  "middleware/auth.go line 42: Missing error handling for jwt.Parse"
  "tests/auth_test.go: No test for Login with invalid password case"

Be direct and precise.`

func (v *ReviewerAgent) Review(ctx context.Context, plan string, modifiedFiles []string, statusCallback StatusCallback) (*ReviewResult, error) {
	if statusCallback != nil {
		statusCallback("Reviewer: Analyzing changes...")
	}

	toolDefs := []interface{}{
		map[string]interface{}{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "read_file",
				"description": "Read file contents",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"path": map[string]interface{}{
							"type":        "string",
							"description": "File path",
						},
					},
					"required": []string{"path"},
				},
			},
		},
		map[string]interface{}{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "run_command",
				"description": "Run shell command to verify build, tests, linter, etc",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"command": map[string]interface{}{
							"type":        "string",
							"description": "Command to execute (e.g. 'go build', 'npm test')",
						},
					},
					"required": []string{"command"},
				},
			},
		},
	}

	filesStr := ""
	if len(modifiedFiles) > 0 {
		filesStr = fmt.Sprintf("\nFiles that were modified: %v", modifiedFiles)
	}

	reviewPrompt := fmt.Sprintf(`Validate if the implementation meets the requirements.

Plan and Success Criteria:
---
%s
---
%s

TASK:
1. Read the modified files to see what was changed
2. **CRITICAL**: Run 'go build' command to verify code compiles
3. Check if changes meet the success criteria from the plan
4. Report:
   - Only say "SUCCESS" if go build exits with code 0 (no errors)
   - If go build fails, report "FAIL" with the specific compilation errors
   - If there are other issues, list them

You MUST run 'go build' for Go projects. Do not skip this step.
Be precise and specific.`, plan, filesStr)

	history := []llm.ChatMessage{
		{Role: "user", Content: reviewPrompt},
	}

	maxIterations := 3
	for i := 0; i < maxIterations; i++ {
		if os.Getenv("CHUCHU_DEBUG") == "1" {
			fmt.Fprintf(os.Stderr, "[VALIDATOR] Iteration %d/%d\n", i+1, maxIterations)
		}

		resp, err := v.provider.Chat(ctx, llm.ChatRequest{
			SystemPrompt: reviewerPrompt,
			Messages:     history,
			Tools:        toolDefs,
			Model:        v.model,
		})
		if err != nil {
			return nil, err
		}

		if len(resp.ToolCalls) == 0 {
			result := &ReviewResult{
				Success:     false,
				Issues:      []string{},
				Suggestions: resp.Text,
			}

			if containsSuccess(resp.Text) {
				result.Success = true
			} else {
				result.Issues = extractIssues(resp.Text)
			}

			if os.Getenv("CHUCHU_DEBUG") == "1" {
				fmt.Fprintf(os.Stderr, "[VALIDATOR] Result: success=%v, issues=%v\n", result.Success, result.Issues)
			}

			return result, nil
		}

		history = append(history, llm.ChatMessage{
			Role:      "assistant",
			Content:   resp.Text,
			ToolCalls: resp.ToolCalls,
		})

		for _, tc := range resp.ToolCalls {
			llmCall := tools.LLMToolCall{
				ID:        tc.ID,
				Name:      tc.Name,
				Arguments: tc.Arguments,
			}
			result := tools.ExecuteToolFromLLM(llmCall, v.cwd)

			content := result.Result
			if result.Error != "" {
				content = "Error: " + result.Error
			}
			if content == "" {
				content = "Success"
			}

			history = append(history, llm.ChatMessage{
				Role:       "tool",
				Content:    content,
				Name:       tc.Name,
				ToolCallID: tc.ID,
			})
		}
	}

	return &ReviewResult{
		Success:     false,
		Issues:      []string{"Validator reached max iterations"},
		Suggestions: "Unable to complete review",
	}, nil
}

func containsSuccess(text string) bool {
	lowerText := strings.ToLower(text)
	// Must explicitly say SUCCESS and not have any failure indicators
	hasSuccess := strings.Contains(lowerText, "success")
	hasFail := strings.Contains(lowerText, "fail") ||
		strings.Contains(lowerText, "error") ||
		strings.Contains(lowerText, "issue") ||
		strings.Contains(lowerText, "problem")
	return hasSuccess && !hasFail
}

func extractIssues(text string) []string {
	issues := []string{}

	lines := strings.Split(text, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		lower := strings.ToLower(trimmed)
		if trimmed != "" && (strings.HasPrefix(trimmed, "-") || strings.HasPrefix(trimmed, "•") || strings.Contains(lower, "issue") || strings.Contains(lower, "missing") || strings.Contains(lower, "error")) {
			issues = append(issues, trimmed)
		}
	}

	if len(issues) == 0 {
		issues = append(issues, text)
	}

	return issues
}
