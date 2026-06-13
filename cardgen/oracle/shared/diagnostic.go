package shared

// Severity is a diagnostic severity.
type Severity uint8

// Diagnostic severities.
const (
	SeverityError Severity = iota + 1
	SeverityWarning
)

// Diagnostic describes a localized source problem.
type Diagnostic struct {
	Severity Severity
	Summary  string
	Detail   string
	Span     Span
}
