package parser

import "testing"

// modifyOrGainExact parses a single creature-modifying sentence (a power/toughness
// pump or a keyword grant) and reports whether its sole resolving effect
// round-tripped to an exact, lowerable production.
func modifyOrGainExact(t *testing.T, source string) bool {
	t.Helper()
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 1 {
		t.Fatalf("Parse(%q) shape = %#v", source, document.Abilities)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) != 1 {
		t.Fatalf("Parse(%q) effects = %#v", source, effects)
	}
	return effects[0].Exact
}

func TestExactGroupKeywordGrantAccepts(t *testing.T) {
	t.Parallel()
	accepted := []string{
		"Creatures you control gain trample until end of turn.",
		"Creatures you control gain first strike and deathtouch until end of turn.",
		"Other creatures you control gain vigilance until end of turn.",
		"Attacking creatures gain first strike until end of turn.",
		"Blocking creatures gain first strike until end of turn.",
		"All creatures gain haste until end of turn.",
	}
	for _, source := range accepted {
		if !modifyOrGainExact(t, source) {
			t.Errorf("modifyOrGainExact(%q) = false, want true", source)
		}
	}
}

func TestExactGroupKeywordGrantFailsClosed(t *testing.T) {
	t.Parallel()
	// Each carries a qualifier or wording the canonical group keyword grant does
	// not reconstruct byte-exactly, so it must not be marked exact.
	rejected := []string{
		"Creatures you control gain protection from red until end of turn.",
		"Creatures you control gain trample.",
		"Creatures you control gain flying until your next turn.",
	}
	for _, source := range rejected {
		if modifyOrGainExact(t, source) {
			t.Errorf("modifyOrGainExact(%q) = true, want false", source)
		}
	}
}

func TestExactAsymmetricDynamicPumpAccepts(t *testing.T) {
	t.Parallel()
	accepted := []string{
		"Target creature gets +X/+X until end of turn, where X is the number of creature cards in your graveyard.",
		"Target creature gets +X/+0 until end of turn, where X is the number of creature cards in your graveyard.",
		"Target creature gets -X/-X until end of turn, where X is the number of cards in your graveyard.",
		"Target creature gets +X/-X until end of turn, where X is the number of cards in your hand.",
	}
	for _, source := range accepted {
		if !modifyOrGainExact(t, source) {
			t.Errorf("modifyOrGainExact(%q) = false, want true", source)
		}
	}
}
