package parser

import "testing"

// sacrificeThenCountEffects parses an instant/sorcery body and returns its
// ordered effects, asserting the body produced exactly one ability with a
// single sentence. The count-scaled sacrifice post-pass annotates the sacrifice
// effect in place rather than fusing the clauses, so callers inspect the
// returned effects directly.
func sacrificeThenCountEffects(t *testing.T, source string) []EffectSyntax {
	t.Helper()
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 || len(document.Abilities) != 1 {
		t.Fatalf("abilities = %#v diagnostics = %#v, want one ability", document.Abilities, diagnostics)
	}
	sentences := document.Abilities[0].Sentences
	if len(sentences) != 1 {
		t.Fatalf("sentences = %#v, want one sentence", sentences)
	}
	return sentences[0].Effects
}

// TestAnnotateSacrificeThenCountAllCreatures proves the parser marks the
// sacrifice effect of "Sacrifice all creatures you control, then create that
// many ..." (Hellion Eruption) as a count-scaled sacrifice that is not an
// any-number choice.
func TestAnnotateSacrificeThenCountAllCreatures(t *testing.T) {
	effects := sacrificeThenCountEffects(t, "Sacrifice all creatures you control, then create that many 4/4 red Hellion creature tokens.")
	if len(effects) != 2 {
		t.Fatalf("effects = %d, want 2 (sacrifice, create)", len(effects))
	}
	sacrifice := effects[0]
	if sacrifice.Kind != EffectSacrifice {
		t.Fatalf("first effect kind = %v, want EffectSacrifice", sacrifice.Kind)
	}
	if !sacrifice.SacrificeThenCount {
		t.Fatal("SacrificeThenCount = false, want true")
	}
	if sacrifice.SacrificeAnyNumber {
		t.Fatal("SacrificeAnyNumber = true, want false for an all-creatures clause")
	}
}

// TestAnnotateSacrificeThenCountAnyNumberLands proves the parser marks the
// sacrifice effect of "Sacrifice any number of lands, then add that much {C}."
// (Mana Seism) as a count-scaled any-number sacrifice.
func TestAnnotateSacrificeThenCountAnyNumberLands(t *testing.T) {
	effects := sacrificeThenCountEffects(t, "Sacrifice any number of lands, then add that much {C}.")
	if len(effects) != 2 {
		t.Fatalf("effects = %d, want 2 (sacrifice, add mana)", len(effects))
	}
	sacrifice := effects[0]
	if sacrifice.Kind != EffectSacrifice {
		t.Fatalf("first effect kind = %v, want EffectSacrifice", sacrifice.Kind)
	}
	if !sacrifice.SacrificeThenCount {
		t.Fatal("SacrificeThenCount = false, want true")
	}
	if !sacrifice.SacrificeAnyNumber {
		t.Fatal("SacrificeAnyNumber = false, want true for an any-number clause")
	}
}

// TestAnnotateSacrificeThenCountIgnoresUnscaledReward proves the post-pass does
// not annotate a sacrifice whose follow-up does not reference the sacrificed
// count, so an ordinary "Sacrifice a creature, then draw a card." stays
// unmarked.
func TestAnnotateSacrificeThenCountIgnoresUnscaledReward(t *testing.T) {
	effects := sacrificeThenCountEffects(t, "Sacrifice a creature, then draw a card.")
	for _, effect := range effects {
		if effect.SacrificeThenCount {
			t.Fatalf("effect %v marked SacrificeThenCount, want unmarked", effect.Kind)
		}
	}
}
