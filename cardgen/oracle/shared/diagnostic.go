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
	// Additional carries other independent blocker reasons discovered on the same
	// failing path, so an unsupported card can report every reason it is blocked
	// rather than only the first. cardgen's fan-out lowerers (ordered sequence,
	// modal, optional) populate it when a construct fails for multiple independent
	// reasons; the card report flattens it into the card's diagnostic list. Reasons
	// travel with the diagnostic, so a speculative lowering whose diagnostic is
	// discarded discards its Additional reasons too. The parser and compiler leave
	// this nil.
	Additional []Diagnostic
}
