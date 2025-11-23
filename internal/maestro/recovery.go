package maestro

import (
	"fmt"
	"strings"
)

// RecoveryStrategy handles error recovery during execution
type RecoveryStrategy struct {
	MaxRetries  int
	Checkpoints *CheckpointSystem
}

// NewRecoveryStrategy creates a new recovery strategy
func NewRecoveryStrategy(maxRetries int, checkpoints *CheckpointSystem) *RecoveryStrategy {
	return &RecoveryStrategy{
		MaxRetries:  maxRetries,
		Checkpoints: checkpoints,
	}
}

// Rollback reverts changes to a specific checkpoint
func (rs *RecoveryStrategy) Rollback(checkpointID string) error {
	if rs.Checkpoints == nil {
		return fmt.Errorf("checkpoint system not initialized")
	}
	return rs.Checkpoints.Restore(checkpointID)
}

// ErrorType represents the category of an error
type ErrorType string

const (
	ErrorSyntax  ErrorType = "syntax"
	ErrorLogic   ErrorType = "logic"
	ErrorBuild   ErrorType = "build"
	ErrorTest    ErrorType = "test"
	ErrorUnknown ErrorType = "unknown"
)

// ClassifyError attempts to categorize an error based on output
func ClassifyError(output string) ErrorType {
	output = strings.ToLower(output)
	if strings.Contains(output, "syntax error") || strings.Contains(output, "expected") || strings.Contains(output, "undefined") || strings.Contains(output, "unexpected") {
		return ErrorSyntax
	}
	if strings.Contains(output, "build failed") || strings.Contains(output, "cannot load") || strings.Contains(output, "compile") {
		return ErrorBuild
	}
	if strings.Contains(output, "test failed") || strings.Contains(output, "fail") || strings.Contains(output, "assert") {
		return ErrorTest
	}
	return ErrorUnknown
}

func (rs *RecoveryStrategy) GenerateFixPrompt(errorType ErrorType, errorOutput string) string {
	var prompt string

	switch errorType {
	case ErrorSyntax:
		prompt = fmt.Sprintf(`The previous implementation has SYNTAX ERRORS.

Error output:
%s

Please fix the syntax errors. Focus on:
- Missing/extra parentheses, brackets, or braces
- Incorrect variable names or typos
- Missing imports or declarations
- Type mismatches

Provide corrected code that compiles without syntax errors.`, errorOutput)

	case ErrorBuild:
		prompt = fmt.Sprintf(`The code failed to BUILD/COMPILE.

Error output:
%s

Please fix the build errors. Common issues:
- Missing dependencies (check imports)
- Type errors or incompatible types
- Undefined functions or variables
- Circular dependencies

Provide corrected code that builds successfully.`, errorOutput)

	case ErrorTest:
		prompt = fmt.Sprintf(`The code builds but TESTS ARE FAILING.

Test output:
%s

Please fix the implementation to make tests pass. Common issues:
- Logic errors in implementation
- Edge cases not handled
- Incorrect return values
- Missing null/error checks

Analyze the failing tests and correct the implementation.`, errorOutput)

	case ErrorLogic:
		prompt = fmt.Sprintf(`The code has LOGIC ERRORS.

Error output:
%s

Please review and fix the logic. Consider:
- Algorithm correctness
- Boundary conditions
- State management
- Data flow

Provide corrected implementation with proper logic.`, errorOutput)

	default:
		prompt = fmt.Sprintf(`The previous implementation FAILED with the following error:

%s

Please analyze the error and provide a corrected implementation.`, errorOutput)
	}

	return prompt
}
