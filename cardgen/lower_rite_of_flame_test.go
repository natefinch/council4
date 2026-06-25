package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
)

// TestLowerRiteOfFlameManaSequence verifies that Rite of Flame lowers to a spell
// ability that adds {R}{R} and then one {R} for each card named Rite of Flame in
// every graveyard (DynamicAmountCardsNamedSourceInGraveyards).
func TestLowerRiteOfFlameManaSequence(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Rite of Flame",
		Layout:     "normal",
		ManaCost:   "{R}",
		TypeLine:   "Sorcery",
		OracleText: "Add {R}{R}, then add {R} for each card named Rite of Flame in each graveyard.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("Rite of Flame did not lower a spell ability")
	}
	content := face.SpellAbility.Val
	if len(content.Modes) != 1 {
		t.Fatalf("modes = %#v, want single mode", content.Modes)
	}
	sequence := content.Modes[0].Sequence
	if len(sequence) != 3 {
		t.Fatalf("sequence = %#v, want three AddMana instructions", sequence)
	}
	for i, instruction := range sequence {
		add, ok := instruction.Primitive.(game.AddMana)
		if !ok || add.ManaColor != mana.R {
			t.Fatalf("instruction %d = %#v, want red AddMana", i, instruction.Primitive)
		}
		switch i {
		case 0, 1:
			if add.Amount.IsDynamic() || add.Amount.Value() != 1 {
				t.Fatalf("instruction %d amount = %#v, want fixed 1", i, add.Amount)
			}
		case 2:
			if !add.Amount.IsDynamic() {
				t.Fatalf("instruction %d amount = %#v, want dynamic count", i, add.Amount)
			}
			dynamic := add.Amount.DynamicAmount().Val
			if dynamic.Kind != game.DynamicAmountCardsNamedSourceInGraveyards ||
				dynamic.Multiplier != 1 {
				t.Fatalf("dynamic amount = %#v", dynamic)
			}
		default:
			t.Fatalf("unexpected instruction index %d", i)
		}
	}
	if err := game.ValidateInstructionSequence(sequence); err != nil {
		t.Fatalf("instruction sequence invalid: %v", err)
	}
}

// TestLowerRiteOfFlameForeignNameFailsClosed verifies that a "for each card
// named <other> in each graveyard" mana tail does not lower when the name is not
// the card's own (the recognizer is text-blind to the card's own name only).
func TestLowerRiteOfFlameForeignNameFailsClosed(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Rite of Flame",
		Layout:     "normal",
		ManaCost:   "{R}",
		TypeLine:   "Sorcery",
		OracleText: "Add {R}{R}, then add {R} for each card named Lightning Bolt in each graveyard.",
	}, "r")
	if err == nil && len(diagnostics) == 0 {
		t.Fatal("expected fail-closed diagnostics for a foreign graveyard name")
	}
}
