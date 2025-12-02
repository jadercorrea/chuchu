package modes

import (
	"context"

	"chuchu/internal/agents"
	"chuchu/internal/autonomous"
	"chuchu/internal/events"
	"chuchu/internal/llm"
)

// AutonomousExecutor wraps autonomous execution for use across modes
type AutonomousExecutor struct {
	events       *events.Emitter
	provider     llm.Provider
	baseProvider llm.Provider
	cwd          string
	model        string
	editorModel  string
	executor     *autonomous.Executor
}

// NewAutonomousExecutor creates a new autonomous executor
func NewAutonomousExecutor(provider llm.Provider, baseProvider llm.Provider, cwd string, model string, editorModel string) *AutonomousExecutor {
	// Create autonomous components
	classifier := agents.NewClassifier(provider, model)
	analyzer := autonomous.NewTaskAnalyzer(classifier, provider, cwd, model)
	planner := agents.NewPlanner(provider, model)
	editor := agents.NewEditor(baseProvider, cwd, editorModel)
	reviewer := agents.NewReviewer(baseProvider, cwd, model)

	executor := autonomous.NewExecutor(analyzer, planner, editor, reviewer, cwd)

	return &AutonomousExecutor{
		events:       events.NewEmitter(nil),
		provider:     provider,
		baseProvider: baseProvider,
		cwd:          cwd,
		model:        model,
		editorModel:  editorModel,
		executor:     executor,
	}
}

// Execute runs autonomous execution with Symphony pattern if task is complex
func (a *AutonomousExecutor) Execute(ctx context.Context, task string) error {
	// Delegate to autonomous executor
	return a.executor.Execute(ctx, task)
}

// ShouldUseAutonomous determines if a task should use autonomous mode
// This is a lightweight heuristic check before full analysis.
// The real complexity scoring happens in TaskAnalyzer.estimateComplexity()
func ShouldUseAutonomous(ctx context.Context, task string) bool {
	// Always return false - let the ML-based complexity analysis decide
	// TaskAnalyzer.Analyze() will trigger Symphony if complexity >= 7
	return false
}
