package parser

import "testing"

func spreeModalFor(t *testing.T, source string) *Modal {
	t.Helper()
	document, diagnostics := Parse(source, Context{CardName: "Test Spree"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %+v; want none", diagnostics)
	}
	for i := range document.Abilities {
		if modal := document.Abilities[i].Modal; modal != nil {
			return modal
		}
	}
	t.Fatalf("no modal ability parsed from %q", source)
	return nil
}

func TestParseSpreeKeyword(t *testing.T) {
	t.Parallel()
	source := "Spree (Choose one or more additional costs.)\n" +
		"+ {1} — Search your library for a card, put it into your graveyard, then shuffle.\n" +
		"+ {2} — Return up to two creature cards with total mana value 4 or less from your graveyard to the battlefield."
	modal := spreeModalFor(t, source)
	if !modal.Spree {
		t.Fatal("modal.Spree = false; want true")
	}
	if !modal.ChoiceKnown || modal.ChoiceKind != ModalChoiceKindOneOrMore {
		t.Fatalf("choice = (known %v, kind %v); want known one-or-more", modal.ChoiceKnown, modal.ChoiceKind)
	}
	if modal.MinModes != 1 || modal.MaxModes != 2 {
		t.Fatalf("modes range = %d/%d; want 1/2", modal.MinModes, modal.MaxModes)
	}
	if len(modal.Options) != 2 {
		t.Fatalf("options = %d; want 2", len(modal.Options))
	}
	for i, want := range []int{1, 2} {
		clause := modal.Options[i].SpreeCost
		if clause == nil {
			t.Fatalf("option %d has no Spree cost clause", i)
		}
		if got := clause.Cost.ManaValue(); got != want {
			t.Fatalf("option %d cost mana value = %d; want %d", i, got, want)
		}
	}
}

func TestParseSpreeKeywordColoredCost(t *testing.T) {
	t.Parallel()
	source := "Spree (Choose one or more additional costs.)\n" +
		"+ {1} — All creatures lose all abilities until end of turn.\n" +
		"+ {3}{W}{W} — Destroy all creatures."
	modal := spreeModalFor(t, source)
	if len(modal.Options) != 2 || modal.MaxModes != 2 {
		t.Fatalf("modal options/max = %d/%d; want 2/2", len(modal.Options), modal.MaxModes)
	}
	second := modal.Options[1].SpreeCost
	if second == nil || second.Cost.ManaValue() != 5 {
		t.Fatalf("second option cost = %+v; want mana value 5", second)
	}
}

// TestParseSpreeKeywordCostlessOptionHasNoCostClause proves the Spree parser
// never fabricates an additional cost: an option printed without a "+ {cost} —"
// clause leaves SpreeCost nil, so the downstream lowering can detect the
// malformed option and fail the card closed (CR 702.171 requires every option
// to carry its own cost).
func TestParseSpreeKeywordCostlessOptionHasNoCostClause(t *testing.T) {
	t.Parallel()
	source := "Spree (Choose one or more additional costs.)\n" +
		"+ {1} — Destroy target artifact.\n" +
		"+ Destroy target enchantment."
	modal := spreeModalFor(t, source)
	if !modal.Spree || len(modal.Options) != 2 {
		t.Fatalf("modal = (spree %v, options %d); want spree with 2 options", modal.Spree, len(modal.Options))
	}
	if modal.Options[0].SpreeCost == nil {
		t.Fatal("first option lost its recognized {1} cost clause")
	}
	if modal.Options[1].SpreeCost != nil {
		t.Fatalf("costless option fabricated a cost clause %+v; want nil", modal.Options[1].SpreeCost)
	}
}

func TestParseGreatTrainHeistSpreeModes(t *testing.T) {
	t.Parallel()
	source := "Spree (Choose one or more additional costs.)\n" +
		"+ {2}{R} — Untap all creatures you control. If it's your combat phase, there is an additional combat phase after this phase.\n" +
		"+ {2} — Creatures you control get +1/+0 and gain first strike until end of turn.\n" +
		"+ {R} — Choose target opponent. Whenever a creature you control deals combat damage to that player this turn, create a tapped Treasure token."
	modal := spreeModalFor(t, source)
	if modal.MinModes != 1 || modal.MaxModes != 3 || len(modal.Options) != 3 {
		t.Fatalf("modal range/options = %d/%d/%d, want 1/3/3", modal.MinModes, modal.MaxModes, len(modal.Options))
	}
	for i, want := range []int{3, 2, 1} {
		if modal.Options[i].SpreeCost == nil || modal.Options[i].SpreeCost.Cost.ManaValue() != want {
			t.Fatalf("mode %d cost = %#v, want mana value %d", i, modal.Options[i].SpreeCost, want)
		}
	}
	var combatCondition bool
	for _, clause := range modal.Options[0].ConditionClauses {
		combatCondition = combatCondition || clause.Predicate == ConditionPredicateControllerCombatPhase
	}
	if !combatCondition {
		t.Fatalf("mode 1 conditions = %#v, want controller combat-phase condition", modal.Options[0].ConditionClauses)
	}
	if len(modal.Options[2].Sentences) != 2 {
		t.Fatalf("mode 3 sentences = %#v, want target declaration and delayed trigger", modal.Options[2].Sentences)
	}
	delayed := modal.Options[2].Sentences[1].Effects
	if len(delayed) != 1 ||
		delayed[0].Kind != EffectDelayedTrigger ||
		!delayed[0].DelayedTriggerBindEventPlayer ||
		delayed[0].DelayedTriggerAbility == nil {
		t.Fatalf("mode 3 delayed effect = %#v", delayed)
	}
}
