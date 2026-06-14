package shared

// Severity is a diagnostic severity.
type Severity string

// Diagnostic severities.
const (
	SeverityError   Severity = "SeverityError"
	SeverityWarning Severity = "SeverityWarning"
)

// Diagnostic describes a localized source problem.
type Diagnostic struct {
	Severity Severity
	Summary  string
	Detail   string
	Span     Span
}
