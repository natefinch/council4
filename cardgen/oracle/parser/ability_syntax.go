package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

// abilitySyntax holds the source-spanned typed structural syntax the parser
// emits for an ability so the compiler consumes it mechanically instead of
// re-deriving the ability's structure from Oracle tokens.
type abilitySyntax struct {
	// bodySpan is the source span of the ability's resolving body: the tokens
	// after the activated/loyalty cost colon or the triggered event comma (and
	// after any ability-word or chapter prefix). Downstream stages select the
	// body tokens structurally by this span instead of re-deriving the boundary
	// from Oracle wording. It is the zero span when the body is empty.
	bodySpan shared.Span
	// cost is the typed cost recognized from the ability's cost phrase.
	cost *Cost
	// optional reports that a triggered ability's resolving body begins with the
	// optional "you may" choice; optionalSpan covers those two words. The parser
	// owns this recognition so the compiler need not inspect "you"/"may" tokens.
	optional     bool
	optionalSpan shared.Span
	// conditionBoundaries lists every condition introducer the parser finds in
	// the ability's condition-scan token stream, in source order, including
	// introducers whose predicate is unrecognized (so the compiler can fail
	// closed). The compiler consumes these typed boundaries instead of scanning
	// Oracle tokens for "if"/"unless"/"only if"/"as long as".
	conditionBoundaries []ConditionBoundary
}

func (a *Ability) ensureStructuralSyntax() *abilitySyntax {
	if a.structuralSyntax == nil {
		a.structuralSyntax = &abilitySyntax{}
	}
	return a.structuralSyntax
}

// BodySpan returns the source span of the ability's resolving body.
func (a *Ability) BodySpan() shared.Span {
	if a.structuralSyntax == nil {
		return shared.Span{}
	}
	return a.structuralSyntax.bodySpan
}

// CostSyntax returns the ability's typed cost, or nil when no cost was parsed.
func (a *Ability) CostSyntax() *Cost {
	if a.structuralSyntax == nil {
		return nil
	}
	return a.structuralSyntax.cost
}

// Optional reports whether a triggered ability's body begins with "you may".
func (a *Ability) Optional() bool {
	if a.structuralSyntax == nil {
		return false
	}
	return a.structuralSyntax.optional
}

// OptionalSpan returns the source span of the leading "you may" choice.
func (a *Ability) OptionalSpan() shared.Span {
	if a.structuralSyntax == nil {
		return shared.Span{}
	}
	return a.structuralSyntax.optionalSpan
}

// ConditionBoundaries returns the ability's typed condition boundaries.
func (a *Ability) ConditionBoundaries() []ConditionBoundary {
	if a.structuralSyntax == nil {
		return nil
	}
	return a.structuralSyntax.conditionBoundaries
}
