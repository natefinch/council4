package cardgen

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
)

func combinationManaAddInstruction(t *testing.T, seq []game.Instruction) game.AddMana {
	t.Helper()
	if len(seq) != 1 {
		t.Fatalf("sequence = %#v, want a single AddMana", seq)
	}
	add, ok := seq[0].Primitive.(game.AddMana)
	if !ok {
		t.Fatalf("primitive = %#v, want AddMana", seq[0].Primitive)
	}
	if err := game.ValidateInstructionSequence(seq); err != nil {
		t.Fatalf("instruction sequence invalid: %v", err)
	}
	return add
}

// TestLowerFixedCombinationMana verifies that fixed "Add <N> mana in any
// combination of <colors>" bodies lower to a single AddMana carrying the fixed
// amount and the offered color set, both for a color subset ("{R} and/or {G}",
// Goblin Clearcutter) and the all-five "colors" wording (Cascading Cataracts).
func TestLowerFixedCombinationMana(t *testing.T) {
	t.Parallel()
	t.Run("subset", func(t *testing.T) {
		t.Parallel()
		face := lowerSingleFace(t, &ScryfallCard{
			Name:       "Goblin Clearcutter",
			Layout:     "normal",
			TypeLine:   "Creature — Goblin",
			ManaCost:   "{3}{R}",
			OracleText: "{T}, Sacrifice a Forest: Add three mana in any combination of {R} and/or {G}.",
		})
		if len(face.ManaAbilities) != 1 {
			t.Fatalf("mana abilities = %d, want 1", len(face.ManaAbilities))
		}
		add := combinationManaAddInstruction(t, face.ManaAbilities[0].Content.Modes[0].Sequence)
		if add.Amount.IsDynamic() || add.Amount.Value() != 3 {
			t.Fatalf("amount = %#v, want fixed 3", add.Amount)
		}
		if !slices.Equal(add.CombinationColors, []mana.Color{mana.R, mana.G}) {
			t.Fatalf("colors = %v, want [R G]", add.CombinationColors)
		}
	})
	t.Run("all colors", func(t *testing.T) {
		t.Parallel()
		face := lowerSingleFace(t, &ScryfallCard{
			Name:       "Cascading Cataracts",
			Layout:     "normal",
			TypeLine:   "Land",
			OracleText: "Indestructible\n{T}: Add {C}.\n{5}, {T}: Add five mana in any combination of colors.",
		})
		if len(face.ManaAbilities) != 2 {
			t.Fatalf("mana abilities = %d, want 2", len(face.ManaAbilities))
		}
		add := combinationManaAddInstruction(t, face.ManaAbilities[1].Content.Modes[0].Sequence)
		if add.Amount.IsDynamic() || add.Amount.Value() != 5 {
			t.Fatalf("amount = %#v, want fixed 5", add.Amount)
		}
		if !slices.Equal(add.CombinationColors, []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G}) {
			t.Fatalf("colors = %v, want all five", add.CombinationColors)
		}
	})
}

// TestLowerDynamicCombinationMana verifies that a dynamic "Add X mana in any
// combination of <colors>, where X is <dynamic>" body lowers to a single AddMana
// carrying the dynamic amount and the offered color set (Burnt Offering, whose X
// is the sacrificed creature's mana value).
func TestLowerDynamicCombinationMana(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Burnt Offering",
		Layout:   "normal",
		TypeLine: "Instant",
		ManaCost: "{B}",
		OracleText: "As an additional cost to cast this spell, sacrifice a creature.\n" +
			"Add X mana in any combination of {B} and/or {R}, where X is the sacrificed creature's mana value.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability missing")
	}
	add := combinationManaAddInstruction(t, face.SpellAbility.Val.Modes[0].Sequence)
	if !add.Amount.IsDynamic() {
		t.Fatalf("amount = %#v, want dynamic", add.Amount)
	}
	if dynamic := add.Amount.DynamicAmount().Val; dynamic.Kind != game.DynamicAmountObjectManaValue {
		t.Fatalf("dynamic amount kind = %v, want object mana value", dynamic.Kind)
	}
	if !slices.Equal(add.CombinationColors, []mana.Color{mana.B, mana.R}) {
		t.Fatalf("colors = %v, want [B R]", add.CombinationColors)
	}
}

// TestLowerCombinationManaRadhaFailsClosed pins Grand Warlord Radha fail-closed:
// its "add that much mana in any combination of {R} and/or {G}" uses an
// unmodeled attacker-count amount and its "you don't lose this mana as steps and
// phases end" rider is unmodeled, so the card must not generate.
func TestLowerCombinationManaRadhaFailsClosed(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:     "Grand Warlord Radha",
		Layout:   "normal",
		TypeLine: "Legendary Creature — Elf Warrior",
		ManaCost: "{2}{R}{G}",
		OracleText: "Haste\n" +
			"Whenever one or more creatures you control attack, add that much mana in any combination of {R} and/or {G}. " +
			"Until end of turn, you don't lose this mana as steps and phases end.",
	}, "x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("Radha unexpectedly generated; it must stay fail-closed")
	}
}
