package parser

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
)

func TestConditionSubtypeListsConsumeAllAlternatives(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name, condition string
		subtypes        []types.Sub
	}{
		{"object list", "If that creature is a Bird, Frog, Otter, or Rat", []types.Sub{"Bird", "Frog", "Otter", "Rat"}},
		{"controls list", "If you control a Fish, Octopus, Otter, Seal, Serpent, or Whale", []types.Sub{"Fish", "Octopus", "Otter", "Seal", "Serpent", "Whale"}},
		{"single subtype", "If that creature is a Bird", []types.Sub{"Bird"}},
		{"unknown first", "If that creature is a Mystery, Frog, Otter, or Rat", nil},
		{"unknown middle", "If that creature is a Bird, Mystery, Otter, or Rat", nil},
		{"unknown last", "If that creature is a Bird, Frog, Otter, or Mystery", nil},
		{"conjunction", "If that creature is a Bird, Frog, Otter, and Rat", nil},
		{"missing member", "If that creature is a Bird, , Otter, or Rat", nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse("Exile target creature, then return it to the battlefield. "+test.condition+", draw a card.", Context{InstantOrSorcery: true})
			clauses := document.Abilities[0].ConditionClauses
			if test.subtypes == nil && len(clauses) == 0 {
				return
			}
			if len(clauses) != 1 {
				t.Fatalf("condition clauses=%d, want one", len(clauses))
			}
			clause := clauses[0]
			recognized := clause.Predicate == ConditionPredicateObjectMatches ||
				clause.Predicate == ConditionPredicateControls
			if test.subtypes == nil {
				if recognized {
					t.Fatalf("malformed list admitted a partial selection: %+v", clause.Selection)
				}
			} else if !recognized || !slices.Equal(clause.Selection.SubtypesAny, test.subtypes) {
				t.Fatalf("selection=%+v predicate=%v, want all %v", clause.Selection, clause.Predicate, test.subtypes)
			}
		})
	}
}

func TestConditionSubtypeNounDoesNotSwallowConsequence(t *testing.T) {
	t.Parallel()
	document, _ := Parse("If you control nine or more Wraiths, Wraiths you control have base power and toughness 9/9 until end of turn.", Context{InstantOrSorcery: true})
	clauses := document.Abilities[0].ConditionClauses
	if len(clauses) != 1 || clauses[0].Predicate != ConditionPredicateControls ||
		!slices.Equal(clauses[0].Selection.SubtypesAny, []types.Sub{"Wraith"}) {
		t.Fatalf("subtype-named consequence changed its condition: %+v", clauses)
	}
}
