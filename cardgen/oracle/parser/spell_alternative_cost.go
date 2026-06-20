package parser

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// SpellAlternativeCostCondition identifies a condition on an alternative spell cost.
type SpellAlternativeCostCondition uint8

// Supported alternative spell-cost conditions.
const (
	SpellAlternativeCostConditionUnknown SpellAlternativeCostCondition = iota
	SpellAlternativeCostConditionControlsCommander
)

// SpellAlternativeCost is typed syntax for a paragraph that offers an
// alternative to the spell's printed mana cost.
type SpellAlternativeCost struct {
	Span                  shared.Span
	Condition             SpellAlternativeCostCondition
	WithoutPayingManaCost bool
}

func spellAlternativeCostClause(body []shared.Token) (*SpellAlternativeCost, bool) {
	if len(body) == 0 || body[len(body)-1].Kind != shared.Period {
		return nil, false
	}
	if !slices.Equal(normalizedWords(body), []string{
		"if", "you", "control", "a", "commander", "you", "may", "cast",
		"this", "spell", "without", "paying", "its", "mana", "cost",
	}) {
		return nil, false
	}
	return &SpellAlternativeCost{
		Span:                  shared.SpanOf(body),
		Condition:             SpellAlternativeCostConditionControlsCommander,
		WithoutPayingManaCost: true,
	}, true
}
