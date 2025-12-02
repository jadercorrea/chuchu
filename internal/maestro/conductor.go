package maestro

import (
	"context"
	"fmt"
	"strings"

	"chuchu/internal/agents"
	"chuchu/internal/llm"
)

// Conductor is the central coordinator (Maestro) that orchestrates all agents
type Conductor struct {
	classifier *agents.Classifier
	planner    *agents.PlannerAgent
	editor     *agents.EditorAgent
	reviewer   *agents.ReviewerAgent
	cwd        string
}

// NewConductor creates a new Maestro conductor
func NewConductor(
	classifier *agents.Classifier,
	planner *agents.PlannerAgent,
	editor *agents.EditorAgent,
	reviewer *agents.ReviewerAgent,
	cwd string,
) *Conductor {
	return &Conductor{
		classifier: classifier,
		planner:    planner,
		editor:     editor,
		reviewer:   reviewer,
		cwd:        cwd,
	}
}

// ExecuteTask orchestrates the execution of a task
func (c *Conductor) ExecuteTask(ctx context.Context, task string) error {
	fmt.Println("Creating plan...")
	plan, err := c.planner.CreatePlan(ctx, task, "", nil)
	if err != nil {
		return fmt.Errorf("planning failed: %w", err)
	}

	// Build conversation history
	history := []llm.ChatMessage{
		{Role: "user", Content: plan},
	}

	maxAttempts := 3
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if attempt > 1 {
			fmt.Printf("Retrying (attempt %d/%d)...\n", attempt, maxAttempts)
		}

		// Execute with editor
		fmt.Println("Executing changes...")
		result, modifiedFiles, err := c.editor.Execute(ctx, history, nil)
		if err != nil {
			if attempt < maxAttempts {
				fmt.Printf("[WARNING] Execution error: %v\n", err)
				feedback := c.formatExecutionError(err)
				history = append(history, llm.ChatMessage{
					Role:    "user",
					Content: feedback,
				})
				continue
			}
			return fmt.Errorf("execution failed: %w", err)
		}

		// Validate
		fmt.Println("Validating...")
		review, err := c.reviewer.Review(ctx, plan, modifiedFiles, nil)
		if err != nil {
			if attempt < maxAttempts {
				fmt.Printf("[WARNING] Validation error: %v\n", err)
				feedback := c.formatValidationError(err)
				history = append(history, llm.ChatMessage{
					Role:    "user",
					Content: feedback,
				})
				continue
			}
			return fmt.Errorf("review failed: %w", err)
		}

		if !review.Success {
			if attempt < maxAttempts {
				issuesStr := strings.Join(review.Issues, "\n")
				fmt.Printf("[WARNING] Validation failed:\n%s\n", issuesStr)
				feedback := c.formatValidationIssues(review.Issues)
				history = append(history, llm.ChatMessage{
					Role:    "user",
					Content: feedback,
				})
				continue
			}
			return fmt.Errorf("review failed after %d attempts: %v", maxAttempts, review.Issues)
		}

		// Success!
		fmt.Printf("\n[OK] Task complete!\n")
		fmt.Printf("   Modified: %d files\n", len(modifiedFiles))
		if result != "" {
			fmt.Printf("   %s\n", result)
		}
		return nil
	}

	return fmt.Errorf("task failed after %d attempts", maxAttempts)
}

// formatExecutionError creates clear feedback for execution errors
func (c *Conductor) formatExecutionError(err error) string {
	return fmt.Sprintf(`EXECUTION FAILED

Error: %v

INSTRUCTIONS:
1. Read the error message carefully
2. Identify what went wrong
3. Fix the specific issue mentioned
4. Try again

Be precise and fix only what's broken.`, err)
}

// formatValidationError creates clear feedback for review errors
func (c *Conductor) formatValidationError(err error) string {
	return fmt.Sprintf(`VALIDATION SYSTEM ERROR

Error: %v

INSTRUCTIONS:
The review process itself failed. This might mean:
- A required tool is not available
- Syntax errors prevent running build/test commands
- File read errors

Please ensure your code is syntactically correct and try again.`, err)
}

// formatValidationIssues creates clear feedback for review failures
func (c *Conductor) formatValidationIssues(issues []string) string {
	var sb strings.Builder
	sb.WriteString("VALIDATION FAILED\n\n")
	sb.WriteString("The following issues were found:\n")
	for i, issue := range issues {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, issue))
	}
	sb.WriteString("\nINSTRUCTIONS:\n")
	sb.WriteString("1. Fix each issue listed above\n")
	sb.WriteString("2. Pay special attention to:\n")

	// Check if there's a Go package mismatch error
	hasPackageError := false
	for _, issue := range issues {
		if strings.Contains(issue, "found packages") {
			hasPackageError = true
			break
		}
	}

	if hasPackageError {
		sb.WriteString("   - **CRITICAL**: Package name mismatch! Read ALL .go files in the directory to see the correct package name, then fix ONLY the wrong file(s)\n")
	}
	sb.WriteString("   - Correct package names (all .go files in same directory must have same package)\n")
	sb.WriteString("   - Missing imports\n")
	sb.WriteString("   - Compilation errors\n")
	sb.WriteString("3. Do NOT change what's already correct\n")
	sb.WriteString("4. Only fix the specific problems mentioned\n")
	return sb.String()
}
