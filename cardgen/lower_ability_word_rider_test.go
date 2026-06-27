package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
)

// TestLowerRallyForTheThroneAdamantRider verifies that the Adamant additive
// rider fuses onto the base spell: an unconditional token creation followed by a
// gain-life instruction gated on three white mana having been spent to cast the
// spell.
func TestLowerRallyForTheThroneAdamantRider(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Rally for the Throne",
		Layout:   "normal",
		TypeLine: "Instant",
		OracleText: "Create two 1/1 white Human creature tokens.\n" +
			"Adamant — If at least three white mana was spent to cast this spell, you gain 1 life for each creature you control.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("Rally for the Throne produced no spell ability")
	}
	modes := face.SpellAbility.Val.Modes
	if len(modes) != 1 {
		t.Fatalf("modes = %d, want 1", len(modes))
	}
	seq := modes[0].Sequence
	if len(seq) != 2 {
		t.Fatalf("sequence length = %d, want 2 (base + rider)", len(seq))
	}
	if _, ok := seq[0].Primitive.(game.CreateToken); !ok {
		t.Fatalf("base instruction = %#v, want CreateToken", seq[0].Primitive)
	}
	if seq[0].Condition.Exists {
		t.Fatal("base token creation must be unconditional")
	}
	if _, ok := seq[1].Primitive.(game.GainLife); !ok {
		t.Fatalf("rider instruction = %#v, want GainLife", seq[1].Primitive)
	}
	if !seq[1].Condition.Exists || !seq[1].Condition.Val.Condition.Exists {
		t.Fatal("rider gain-life must be gated")
	}
	gate := seq[1].Condition.Val.Condition.Val
	if gate.SpellColorManaSpent.Color != color.White || gate.SpellColorManaSpent.Count != 3 {
		t.Fatalf("rider gate = %+v, want three white mana spent", gate.SpellColorManaSpent)
	}
	if gate.Negate {
		t.Fatal("rider gate must not be negated")
	}
}

// TestLowerUnexplainedVisionAdamantRider verifies an untargeted base draw with
// an untargeted Adamant scry rider fuses into one spell.
func TestLowerUnexplainedVisionAdamantRider(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Unexplained Vision",
		Layout:   "normal",
		TypeLine: "Sorcery",
		OracleText: "Draw three cards.\n" +
			"Adamant — If at least three blue mana was spent to cast this spell, scry 3.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("Unexplained Vision produced no spell ability")
	}
	seq := face.SpellAbility.Val.Modes[0].Sequence
	if len(seq) != 2 {
		t.Fatalf("sequence length = %d, want 2", len(seq))
	}
	if _, ok := seq[0].Primitive.(game.Draw); !ok {
		t.Fatalf("base instruction = %#v, want Draw", seq[0].Primitive)
	}
	if seq[0].Condition.Exists {
		t.Fatal("base draw must be unconditional")
	}
	if _, ok := seq[1].Primitive.(game.Scry); !ok {
		t.Fatalf("rider instruction = %#v, want Scry", seq[1].Primitive)
	}
	if !seq[1].Condition.Exists {
		t.Fatal("rider scry must be gated on the Adamant condition")
	}
}

// TestLowerForebodingFruitTargetedBaseAdamantRider verifies the rider appends to
// a targeted base spell: the base's target player draws and loses life, then the
// Adamant Food-token creation resolves only when three black mana was spent. The
// fused spell keeps the base's single target.
func TestLowerForebodingFruitTargetedBaseAdamantRider(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Foreboding Fruit",
		Layout:   "normal",
		TypeLine: "Sorcery",
		OracleText: "Target player draws two cards and loses 2 life.\n" +
			"Adamant — If at least three black mana was spent to cast this spell, create a Food token. (It's an artifact with \"{2}, {T}, Sacrifice this token: You gain 3 life.\")",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("Foreboding Fruit produced no spell ability")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("mode targets = %d, want 1 (target player)", len(mode.Targets))
	}
	seq := mode.Sequence
	if len(seq) != 3 {
		t.Fatalf("sequence length = %d, want 3 (draw, lose, rider)", len(seq))
	}
	if !seq[len(seq)-1].Condition.Exists {
		t.Fatal("rider must be the gated trailing instruction")
	}
	if _, ok := seq[len(seq)-1].Primitive.(game.CreateToken); !ok {
		t.Fatalf("rider instruction = %#v, want CreateToken", seq[len(seq)-1].Primitive)
	}
}

// TestRejectNonRulesFreeAbilityWordRider confirms the combiner fails closed when
// the rider paragraph's ability word is not a recognized rules-free label, so an
// unrecognized label is never silently dropped.
func TestRejectNonRulesFreeAbilityWordRider(t *testing.T) {
	t.Parallel()
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:     "Fake Rider",
		Layout:   "normal",
		TypeLine: "Sorcery",
		OracleText: "Draw three cards.\n" +
			"Strive — If at least three blue mana was spent to cast this spell, scry 3.",
	})
}
