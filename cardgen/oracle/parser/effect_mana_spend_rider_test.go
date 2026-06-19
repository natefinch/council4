package parser

import (
	"testing"
)

// manaSpendRiderAbility wraps the rider sentence in the Path of Ancestry mana
// ability so the parser processes it in the same activated-ability context as
// the real card.
func manaSpendRiderAbility(rider string) string {
	return "{T}: Add one mana of any color in your commander's color identity. " + rider
}

// riderEffect returns the lone rider effect of the parsed mana ability, or nil
// when the sentence did not collapse to a single EffectManaSpendRider.
func riderEffect(t *testing.T, src string) *EffectSyntax {
	t.Helper()
	document, diagnostics := Parse(src, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 1 {
		t.Fatalf("abilities = %d, want 1", len(document.Abilities))
	}
	for si := range document.Abilities[0].Sentences {
		sentence := &document.Abilities[0].Sentences[si]
		for ei := range sentence.Effects {
			if sentence.Effects[ei].Kind == EffectManaSpendRider {
				return &sentence.Effects[ei]
			}
		}
	}
	return nil
}

// TestParseManaSpendRiderExact verifies that Path of Ancestry's mana-spend rider
// collapses into a single typed EffectManaSpendRider with the commander
// creature-type condition and the scry effect.
func TestParseManaSpendRiderExact(t *testing.T) {
	t.Parallel()
	effect := riderEffect(t, manaSpendRiderAbility(
		"When that mana is spent to cast a creature spell that shares a creature type with your commander, scry 1.",
	))
	if effect == nil {
		t.Fatal("rider sentence did not collapse to EffectManaSpendRider")
	}
	if effect.ManaSpendRider == nil {
		t.Fatalf("ManaSpendRider = nil, effect = %#v", effect)
	}
	if effect.ManaSpendRider.Condition != ManaSpendCastCommanderCreatureType {
		t.Fatalf("Condition = %q, want %q", effect.ManaSpendRider.Condition, ManaSpendCastCommanderCreatureType)
	}
	if effect.ManaSpendRider.Effect != ManaSpendRiderEffectScry {
		t.Fatalf("Effect = %q, want %q", effect.ManaSpendRider.Effect, ManaSpendRiderEffectScry)
	}
	if effect.ManaSpendRider.ScryAmount != 1 {
		t.Fatalf("ScryAmount = %d, want 1", effect.ManaSpendRider.ScryAmount)
	}
	if !effect.Exact {
		t.Fatal("Exact = false, want true")
	}
}

// TestParseManaSpendRiderScryAmount verifies the scry amount is read from the
// integer token rather than assumed to be 1.
func TestParseManaSpendRiderScryAmount(t *testing.T) {
	t.Parallel()
	effect := riderEffect(t, manaSpendRiderAbility(
		"When that mana is spent to cast a creature spell that shares a creature type with your commander, scry 2.",
	))
	if effect == nil || effect.ManaSpendRider == nil {
		t.Fatal("rider sentence did not collapse to EffectManaSpendRider")
	}
	if effect.ManaSpendRider.ScryAmount != 2 {
		t.Fatalf("ScryAmount = %d, want 2", effect.ManaSpendRider.ScryAmount)
	}
}

// TestParseManaSpendRiderFailClosed verifies that near-miss riders never
// collapse to EffectManaSpendRider, so a different spend condition, an
// unrestricted "when this mana is spent", a non-creature spell qualifier, or a
// different rider effect all fall back to generic effects that fail closed in
// the compiler and lowering.
func TestParseManaSpendRiderFailClosed(t *testing.T) {
	t.Parallel()
	for _, rider := range []string{
		// Unrestricted "when this mana is spent" (Pyromancer's Goggles shape).
		"When that mana is spent to cast a spell, scry 1.",
		// Different spell qualifier (any spell, not a creature spell).
		"When that mana is spent to cast a creature spell, scry 1.",
		// "this mana" rather than "that mana".
		"When this mana is spent to cast a creature spell that shares a creature type with your commander, scry 1.",
		// Different rider effect (draw rather than scry).
		"When that mana is spent to cast a creature spell that shares a creature type with your commander, draw a card.",
		// Trailing unmodeled qualifier after the scry amount.
		"When that mana is spent to cast a creature spell that shares a creature type with your commander, scry 1 then draw a card.",
		// Scry zero is not a positive scry.
		"When that mana is spent to cast a creature spell that shares a creature type with your commander, scry 0.",
		// Shares a type with a different object than the commander.
		"When that mana is spent to cast a creature spell that shares a creature type with another creature you control, scry 1.",
	} {
		document, _ := Parse(manaSpendRiderAbility(rider), Context{})
		if len(document.Abilities) != 1 {
			continue
		}
		for si := range document.Abilities[0].Sentences {
			for ei := range document.Abilities[0].Sentences[si].Effects {
				if document.Abilities[0].Sentences[si].Effects[ei].Kind == EffectManaSpendRider {
					t.Fatalf("near-miss rider %q collapsed to EffectManaSpendRider", rider)
				}
			}
		}
	}
}
