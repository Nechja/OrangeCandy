package snapshot

type Diagnosis struct {
	Exception       *ExceptionInfo `json:"exception"`
	Locals          []Local        `json:"locals_at_crash"`
	CallStack       []Frame        `json:"call_stack"`
	SourceLines     []string       `json:"source_window"`
	CrashLocation   Location       `json:"crash_location"`
	TimelineContext string         `json:"timeline_context"`
}

type ExceptionInfo struct {
	Type       string          `json:"type"`
	Message    string          `json:"message"`
	StackTrace string          `json:"stack_trace,omitempty"`
	Inner      []ExceptionInfo `json:"inner,omitempty"`
}
