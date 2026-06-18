package parser

import "testing"

// graveyardReturnExact parses a single graveyard-return sentence and reports
// whether its resolving effect round-tripped to an exact, lowerable production.
func graveyardReturnExact(t *testing.T, source string) bool {
	t.Helper()
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 1 {
		t.Fatalf("Parse(%q) shape = %#v", source, document.Abilities)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) != 1 || effects[0].Kind != EffectReturn {
		t.Fatalf("Parse(%q) effects = %#v", source, effects)
	}
	return effects[0].Exact
}

func TestExactGraveyardCardTargetAccepts(t *testing.T) {
	t.Parallel()
	accepted := []string{
		"Return target card from your graveyard to your hand.",
		"Return target creature card from your graveyard to your hand.",
		"Return target artifact card from an opponent's graveyard to your hand.",
		"Return target land card from a graveyard to your hand.",
		"Return target creature or enchantment card from your graveyard to your hand.",
		"Return target artifact or creature card from your graveyard to the battlefield.",
		"Return target instant or sorcery card from your graveyard to your hand.",
		"Return target permanent card from your graveyard to the battlefield.",
		"Return target green card from your graveyard to your hand.",
		"Return target multicolored card from your graveyard to your hand.",
		"Return target colorless card from your graveyard to your hand.",
		"Return target Zombie card from your graveyard to the battlefield.",
		"Return target creature card with mana value 3 or less from your graveyard to your hand.",
		"Return another target creature card from your graveyard to your hand.",
		"Return two target creature cards from your graveyard to your hand.",
		"Return up to two target creature cards from your graveyard to your hand.",
		"Return up to three target permanent cards from your graveyard to your hand.",
		"Return up to two target Zombie cards from your graveyard to the battlefield.",
		"Return up to two target cards with cycling from your graveyard to your hand.",
	}
	for _, source := range accepted {
		if !graveyardReturnExact(t, source) {
			t.Errorf("graveyardReturnExact(%q) = false, want true", source)
		}
	}
}

func TestExactGraveyardCardTargetFailsClosed(t *testing.T) {
	t.Parallel()
	// Each of these carries a qualifier the canonical graveyard-card phrasing
	// cannot faithfully reconstruct, so the round-trip must fail closed and the
	// card must keep failing rather than lower to a wrong predicate.
	rejected := []string{
		// Single instant/sorcery types are not retained by the compiler's
		// single-type path, so they would lower to an unrestricted card.
		"Return target sorcery card from your graveyard to your hand.",
		"Return target instant card from your graveyard to your hand.",
		// Supertype, excluded type, and color+type combinations are unrendered.
		"Return target basic land card from your graveyard to the battlefield.",
		"Return target nonland permanent card from your graveyard to the battlefield.",
		"Return target blue creature card from your graveyard to your hand.",
		// "and/or" unions are not the canonical " or " join.
		"Return up to two target instant and/or sorcery cards from your graveyard to your hand.",
	}
	for _, source := range rejected {
		if graveyardReturnExact(t, source) {
			t.Errorf("graveyardReturnExact(%q) = true, want false (fail closed)", source)
		}
	}
}
