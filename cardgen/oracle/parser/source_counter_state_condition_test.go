package parser

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/counter"
)

// TestParseKeywordGrantGatedOnSelfCounterIsSingleEffect proves the keyword grant
// "This creature has trample as long as it has a +1/+1 counter on it." parses to
// one effect. The gating clause's "has a +1/+1 counter" possession verb must be
// reclassified away from a keyword grant so it does not segment a spurious second
// effect that would strand the static declaration as a mixed keyword ability.
func TestParseKeywordGrantGatedOnSelfCounterIsSingleEffect(t *testing.T) {
	t.Parallel()
	ability := parseSingleAbility(t, "This creature has trample as long as it has a +1/+1 counter on it.", Context{})
	if len(ability.Sentences) != 1 {
		t.Fatalf("sentences = %d, want one", len(ability.Sentences))
	}
	if got := len(ability.Sentences[0].Effects); got != 1 {
		t.Fatalf("effects = %d, want one keyword grant (counter-possession gate must not segment a duplicate effect)", got)
	}
	if ability.Sentences[0].Effects[0].Kind != EffectGrantKeyword {
		t.Fatalf("effect kind = %v, want keyword grant", ability.Sentences[0].Effects[0].Kind)
	}
}

// self-counter condition "this creature has a +1/+1 counter on it" parses to one
// source object-match clause requiring at least one +1/+1 counter, the gate
// Incubation Druid's mana multiplier rider reads. The singular "a <kind> counter"
// presence means one or more counters of that kind.
func TestParseSingularSelfCounterStateCondition(t *testing.T) {
	t.Parallel()
	clause := parseSingleConditionClause(t, "this creature has a +1/+1 counter on it")
	if clause.Predicate != ConditionPredicateObjectMatches {
		t.Fatalf("predicate = %v, want object-matches", clause.Predicate)
	}
	if clause.ObjectBinding != ConditionObjectBindingSource {
		t.Fatalf("object binding = %v, want source", clause.ObjectBinding)
	}
	if !clause.Selection.CounterKindKnown ||
		clause.Selection.CounterKind != counter.PlusOnePlusOne ||
		clause.Selection.CounterCountAtLeast != 1 {
		t.Fatalf("selection = %#v, want +1/+1 count >= 1", clause.Selection)
	}
}

// self-counter condition "this creature has fewer than three +1/+1 counters on
// it" parses to one source object-match clause carrying a strict upper bound of
// three, the intervening-if gate Runaway Steam-Kin reads to stop growing at three
// +1/+1 counters. The strict bound sets CounterCountLessThan and leaves the
// inclusive minimum zero so the two thresholds stay mutually exclusive.
func TestParseFewerThanSelfCounterStateCondition(t *testing.T) {
	t.Parallel()
	for _, wording := range []string{
		"this creature has fewer than three +1/+1 counters on it",
		"this creature has less than three +1/+1 counters on it",
	} {
		clause := parseSingleConditionClause(t, wording)
		if clause.Predicate != ConditionPredicateObjectMatches {
			t.Fatalf("%q predicate = %v, want object-matches", wording, clause.Predicate)
		}
		if clause.ObjectBinding != ConditionObjectBindingSource {
			t.Fatalf("%q object binding = %v, want source", wording, clause.ObjectBinding)
		}
		if !clause.Selection.CounterKindKnown ||
			clause.Selection.CounterKind != counter.PlusOnePlusOne ||
			clause.Selection.CounterCountLessThan != 3 ||
			clause.Selection.CounterCountAtLeast != 0 {
			t.Fatalf("%q selection = %#v, want +1/+1 count < 3", wording, clause.Selection)
		}
	}
}
