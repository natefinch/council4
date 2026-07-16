package parser

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/counter"
)

// TestParseSourceThresholdCounterCondition proves the positive source-counter
// threshold "there are three or more ribbon counters on this creature" parses to
// one source object-match clause requiring at least three ribbon counters, the
// gate Prize Pig reads before removing those counters and untapping itself. It is
// the affirmative counterpart of the "there are no <kind> counters on this
// <type>" recognizer.
func TestParseSourceThresholdCounterCondition(t *testing.T) {
	t.Parallel()
	clause := parseSingleConditionClause(t, "there are three or more ribbon counters on this creature")
	if clause.Predicate != ConditionPredicateObjectMatches {
		t.Fatalf("predicate = %v, want object-matches", clause.Predicate)
	}
	if clause.ObjectBinding != ConditionObjectBindingSource {
		t.Fatalf("object binding = %v, want source", clause.ObjectBinding)
	}
	if clause.Negated {
		t.Fatal("clause negated, want the positive threshold form")
	}
	if !clause.Selection.CounterKindKnown ||
		clause.Selection.CounterKind != counter.Ribbon ||
		clause.Selection.CounterCountAtLeast != 3 ||
		clause.Selection.CounterCountLessThan != 0 {
		t.Fatalf("selection = %#v, want ribbon count >= 3", clause.Selection)
	}
}

// TestParseRemoveThoseCountersSegmentsEffect proves the sequence tail "Then if
// there are three or more ribbon counters on this creature, remove those counters
// and untap it." segments into two effects — a back-referencing counter removal
// ("remove those counters") and an untap — rather than collapsing the whole
// clause into a single untap. The removal names no explicit "from" clause, so it
// is recognized through the "those counters" anchor and marked RemoveThoseCounters
// for kind resolution during lowering.
func TestParseRemoveThoseCountersSegmentsEffect(t *testing.T) {
	t.Parallel()
	ability := parseSingleAbility(t,
		"Whenever you gain life, put that many ribbon counters on this creature. Then if there are three or more ribbon counters on this creature, remove those counters and untap it.",
		Context{})
	if len(ability.Sentences) != 2 {
		t.Fatalf("sentences = %d, want 2", len(ability.Sentences))
	}
	tail := ability.Sentences[1].Effects
	if len(tail) != 2 {
		t.Fatalf("tail effects = %d, want 2 (remove those counters, untap)", len(tail))
	}
	if tail[0].Kind != EffectRemoveCounter || !tail[0].RemoveThoseCounters {
		t.Fatalf("tail[0] = kind %v RemoveThoseCounters %v, want remove-counter marked RemoveThoseCounters",
			tail[0].Kind, tail[0].RemoveThoseCounters)
	}
	if tail[1].Kind != EffectUntap {
		t.Fatalf("tail[1] kind = %v, want untap", tail[1].Kind)
	}
}
