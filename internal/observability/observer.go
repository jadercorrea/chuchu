package observability

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// Event is the base interface for all observer events
type Event interface {
	EventType() string
	Timestamp() time.Time
}

// BaseEvent provides common fields for all events
type BaseEvent struct {
	Time time.Time `json:"timestamp"`
}

func (e BaseEvent) Timestamp() time.Time { return e.Time }

// ToolCallEvent is emitted when a tool is called
type ToolCallEvent struct {
	BaseEvent
	Name      string        `json:"name"`
	Arguments string        `json:"arguments"`
	Result    string        `json:"result"`
	Duration  time.Duration `json:"duration_ms"`
	Error     string        `json:"error,omitempty"`
}

func (e ToolCallEvent) EventType() string { return "tool_call" }

// FileModifiedEvent is emitted when a file is created, modified, or deleted
type FileModifiedEvent struct {
	BaseEvent
	Path      string `json:"path"`
	Operation string `json:"operation"` // "create", "modify", "delete"
	Bytes     int64  `json:"bytes"`
}

func (e FileModifiedEvent) EventType() string { return "file_modified" }

// LLMRequestEvent is emitted for each LLM API call
type LLMRequestEvent struct {
	BaseEvent
	Model     string        `json:"model"`
	Backend   string        `json:"backend"`
	TokensIn  int           `json:"tokens_in"`
	TokensOut int           `json:"tokens_out"`
	Duration  time.Duration `json:"duration_ms"`
	Error     string        `json:"error,omitempty"`
}

func (e LLMRequestEvent) EventType() string { return "llm_request" }

// AgentEvent is emitted when an agent starts or ends
type AgentEvent struct {
	BaseEvent
	Name    string `json:"name"`
	Phase   string `json:"phase"` // "start", "end"
	Success bool   `json:"success,omitempty"`
	Message string `json:"message,omitempty"`
}

func (e AgentEvent) EventType() string { return "agent" }

// MovementEvent is emitted for Symphony movements
type MovementEvent struct {
	BaseEvent
	ID      string `json:"id"`
	Name    string `json:"name"`
	Phase   string `json:"phase"` // "start", "end"
	Success bool   `json:"success,omitempty"`
}

func (e MovementEvent) EventType() string { return "movement" }

// ValidationEvent is emitted after validation
type ValidationEvent struct {
	BaseEvent
	Success bool     `json:"success"`
	Issues  []string `json:"issues,omitempty"`
}

func (e ValidationEvent) EventType() string { return "validation" }

// ExecutionSummary provides a summary of the entire execution
type ExecutionSummary struct {
	Duration      time.Duration     `json:"duration_ms"`
	FilesCreated  []string          `json:"files_created"`
	FilesModified []string          `json:"files_modified"`
	FilesDeleted  []string          `json:"files_deleted"`
	ToolCalls     map[string]int    `json:"tool_calls"`
	LLMCalls      int               `json:"llm_calls"`
	TokensIn      int               `json:"tokens_in"`
	TokensOut     int               `json:"tokens_out"`
	Errors        []string          `json:"errors"`
	Success       bool              `json:"success"`
}

// Observer is the interface for tracking agent activity
type Observer interface {
	// Emit records an event
	Emit(event Event)

	// Summary returns the execution summary
	Summary() *ExecutionSummary

	// Subscribe allows real-time event streaming (for LiveDashboard)
	Subscribe(ch chan<- Event)

	// Unsubscribe removes a subscriber
	Unsubscribe(ch chan<- Event)

	// SetVerbose enables real-time console output
	SetVerbose(verbose bool)
}

// AgentObserver is the concrete implementation of Observer
type AgentObserver struct {
	mu          sync.RWMutex
	startTime   time.Time
	events      []Event
	subscribers []chan<- Event
	verbose     bool

	// Aggregated stats
	filesCreated  map[string]int64 // path -> bytes
	filesModified map[string]int64
	filesDeleted  []string
	toolCalls     map[string]int
	llmCalls      int
	tokensIn      int
	tokensOut     int
	errors        []string
	success       bool
}

// NewObserver creates a new AgentObserver
func NewObserver() *AgentObserver {
	return &AgentObserver{
		startTime:     time.Now(),
		events:        make([]Event, 0),
		subscribers:   make([]chan<- Event, 0),
		filesCreated:  make(map[string]int64),
		filesModified: make(map[string]int64),
		filesDeleted:  make([]string, 0),
		toolCalls:     make(map[string]int),
		errors:        make([]string, 0),
		success:       true,
	}
}

// Emit records an event and notifies subscribers
func (o *AgentObserver) Emit(event Event) {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.events = append(o.events, event)

	// Update aggregated stats based on event type
	switch e := event.(type) {
	case *ToolCallEvent:
		o.toolCalls[e.Name]++
		if e.Error != "" {
			o.errors = append(o.errors, e.Error)
		}
	case *FileModifiedEvent:
		switch e.Operation {
		case "create":
			o.filesCreated[e.Path] = e.Bytes
		case "modify":
			o.filesModified[e.Path] = e.Bytes
		case "delete":
			o.filesDeleted = append(o.filesDeleted, e.Path)
		}
	case *LLMRequestEvent:
		o.llmCalls++
		o.tokensIn += e.TokensIn
		o.tokensOut += e.TokensOut
		if e.Error != "" {
			o.errors = append(o.errors, e.Error)
		}
	case *AgentEvent:
		if e.Phase == "end" && !e.Success {
			o.success = false
		}
	case *ValidationEvent:
		if !e.Success {
			o.errors = append(o.errors, e.Issues...)
		}
	}

	// Notify subscribers (non-blocking)
	for _, ch := range o.subscribers {
		select {
		case ch <- event:
		default:
			// Drop if subscriber is slow
		}
	}

	// Verbose mode: print to console
	if o.verbose {
		o.printEvent(event)
	}
}

// printEvent outputs event details to console
func (o *AgentObserver) printEvent(event Event) {
	switch e := event.(type) {
	case *ToolCallEvent:
		if e.Error != "" {
			fmt.Printf("  [TOOL] %s → ERROR: %s\n", e.Name, e.Error)
		} else {
			result := e.Result
			if len(result) > 80 {
				result = result[:80] + "..."
			}
			fmt.Printf("  [TOOL] %s (%.1fs)\n", e.Name, e.Duration.Seconds())
		}
	case *FileModifiedEvent:
		switch e.Operation {
		case "create":
			fmt.Printf("  [FILE] + %s (%d bytes)\n", e.Path, e.Bytes)
		case "modify":
			fmt.Printf("  [FILE] ~ %s (%d bytes)\n", e.Path, e.Bytes)
		case "delete":
			fmt.Printf("  [FILE] - %s\n", e.Path)
		}
	case *LLMRequestEvent:
		fmt.Printf("  [LLM] %s: %d→%d tokens (%.1fs)\n",
			e.Model, e.TokensIn, e.TokensOut, e.Duration.Seconds())
	case *AgentEvent:
		if e.Phase == "start" {
			fmt.Printf("  [AGENT] %s started\n", e.Name)
		} else {
			status := "✓"
			if !e.Success {
				status = "✗"
			}
			fmt.Printf("  [AGENT] %s ended %s\n", e.Name, status)
		}
	case *MovementEvent:
		if e.Phase == "start" {
			fmt.Printf("\n→ Movement: %s\n", e.Name)
		} else {
			status := "✓"
			if !e.Success {
				status = "✗"
			}
			fmt.Printf("← Movement complete %s\n", status)
		}
	}
}

// Summary returns the execution summary
func (o *AgentObserver) Summary() *ExecutionSummary {
	o.mu.RLock()
	defer o.mu.RUnlock()

	created := make([]string, 0, len(o.filesCreated))
	for path := range o.filesCreated {
		created = append(created, path)
	}

	modified := make([]string, 0, len(o.filesModified))
	for path := range o.filesModified {
		modified = append(modified, path)
	}

	return &ExecutionSummary{
		Duration:      time.Since(o.startTime),
		FilesCreated:  created,
		FilesModified: modified,
		FilesDeleted:  o.filesDeleted,
		ToolCalls:     o.toolCalls,
		LLMCalls:      o.llmCalls,
		TokensIn:      o.tokensIn,
		TokensOut:     o.tokensOut,
		Errors:        o.errors,
		Success:       o.success && len(o.errors) == 0,
	}
}

// Subscribe adds a channel to receive events
func (o *AgentObserver) Subscribe(ch chan<- Event) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.subscribers = append(o.subscribers, ch)
}

// Unsubscribe removes a subscriber
func (o *AgentObserver) Unsubscribe(ch chan<- Event) {
	o.mu.Lock()
	defer o.mu.Unlock()
	for i, sub := range o.subscribers {
		if sub == ch {
			o.subscribers = append(o.subscribers[:i], o.subscribers[i+1:]...)
			break
		}
	}
}

// SetVerbose enables or disables real-time console output
func (o *AgentObserver) SetVerbose(verbose bool) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.verbose = verbose
}

// Events returns all recorded events
func (o *AgentObserver) Events() []Event {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return append([]Event{}, o.events...)
}

// GetFilesCreated returns the list of created files
func (o *AgentObserver) GetFilesCreated() []string {
	o.mu.RLock()
	defer o.mu.RUnlock()
	result := make([]string, 0, len(o.filesCreated))
	for path := range o.filesCreated {
		result = append(result, path)
	}
	return result
}

// GetFilesModified returns the list of modified files
func (o *AgentObserver) GetFilesModified() []string {
	o.mu.RLock()
	defer o.mu.RUnlock()
	result := make([]string, 0, len(o.filesModified))
	for path := range o.filesModified {
		result = append(result, path)
	}
	return result
}

// AllModifiedFiles returns created + modified files (for backward compat)
func (o *AgentObserver) AllModifiedFiles() []string {
	o.mu.RLock()
	defer o.mu.RUnlock()
	result := make([]string, 0, len(o.filesCreated)+len(o.filesModified))
	for path := range o.filesCreated {
		result = append(result, path)
	}
	for path := range o.filesModified {
		result = append(result, path)
	}
	return result
}

// PrintSummary outputs a formatted summary to console
func (o *AgentObserver) PrintSummary() {
	summary := o.Summary()

	fmt.Println("\n" + strings.Repeat("─", 50))
	fmt.Println("📊 EXECUTION SUMMARY")
	fmt.Println(strings.Repeat("─", 50))

	// Duration
	fmt.Printf("⏱️  Duration: %.1fs\n", summary.Duration.Seconds())

	// Files
	totalFiles := len(summary.FilesCreated) + len(summary.FilesModified)
	if totalFiles > 0 {
		fmt.Printf("\n📁 Files Changed: %d\n", totalFiles)
		for _, f := range summary.FilesCreated {
			bytes := o.filesCreated[f]
			fmt.Printf("   ✚ %s (%d bytes)\n", f, bytes)
		}
		for _, f := range summary.FilesModified {
			bytes := o.filesModified[f]
			fmt.Printf("   ✎ %s (%d bytes)\n", f, bytes)
		}
		for _, f := range summary.FilesDeleted {
			fmt.Printf("   ✖ %s\n", f)
		}
	} else {
		fmt.Println("\n📁 Files Changed: 0")
	}

	// Tool calls
	if len(summary.ToolCalls) > 0 {
		fmt.Println("\n🔧 Tool Calls:")
		for tool, count := range summary.ToolCalls {
			fmt.Printf("   • %s: %d\n", tool, count)
		}
	}

	// LLM stats
	if summary.LLMCalls > 0 {
		fmt.Printf("\n🤖 LLM: %d calls, %d tokens in, %d tokens out\n",
			summary.LLMCalls, summary.TokensIn, summary.TokensOut)
	}

	// Errors
	if len(summary.Errors) > 0 {
		fmt.Printf("\n⚠️  Errors: %d\n", len(summary.Errors))
		for _, err := range summary.Errors {
			fmt.Printf("   • %s\n", err)
		}
	}

	// Final status
	fmt.Println()
	if summary.Success {
		fmt.Println("✅ Status: SUCCESS")
	} else {
		fmt.Println("❌ Status: FAILED")
	}
	fmt.Println(strings.Repeat("─", 50))
}
