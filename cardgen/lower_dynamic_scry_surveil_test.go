package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerDynamicWhereXSurveilSpell proves a standalone surveil whose amount is
// a "where X is <count>" clause lowers to game.Surveil with a dynamic
// count-selector quantity rather than failing closed as it did when the
// controller scry/surveil path accepted only a fixed literal amount.
func TestLowerDynamicWhereXSurveilSpell(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Surveil X",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Surveil X, where X is the number of artifacts you control.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence = %d, want 1", len(mode.Sequence))
	}
	surveil, ok := mode.Sequence[0].Primitive.(game.Surveil)
	if !ok {
		t.Fatalf("primitive = %T, want game.Surveil", mode.Sequence[0].Primitive)
	}
	if surveil.Player != game.ControllerReference() {
		t.Fatalf("surveil.Player = %+v, want controller", surveil.Player)
	}
	dyn := surveil.Amount.DynamicAmount()
	if !dyn.Exists {
		t.Fatalf("surveil.Amount = %+v, want a dynamic amount", surveil.Amount)
	}
	if dyn.Val.Kind != game.DynamicAmountCountSelector {
		t.Fatalf("dynamic kind = %v, want DynamicAmountCountSelector", dyn.Val.Kind)
	}
}

// TestLowerDynamicWhereXScrySequence proves a scry whose amount is the greatest
// mana value among permanents you control lowers inside an ordered sequence
// ahead of a fixed draw, the form Ugin's Insight uses.
func TestLowerDynamicWhereXScrySequence(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Insight",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Scry X, where X is the greatest mana value among permanents you control, then draw three cards.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %d, want 2", len(mode.Sequence))
	}
	scry, ok := mode.Sequence[0].Primitive.(game.Scry)
	if !ok {
		t.Fatalf("primitive[0] = %T, want game.Scry", mode.Sequence[0].Primitive)
	}
	dyn := scry.Amount.DynamicAmount()
	if !dyn.Exists || dyn.Val.Kind != game.DynamicAmountGreatestManaValueInGroup {
		t.Fatalf("scry.Amount = %+v, want dynamic greatest-mana-value", scry.Amount)
	}
	draw, ok := mode.Sequence[1].Primitive.(game.Draw)
	if !ok {
		t.Fatalf("primitive[1] = %T, want game.Draw", mode.Sequence[1].Primitive)
	}
	if draw.Amount.Value() != 3 {
		t.Fatalf("draw.Amount = %d, want 3", draw.Amount.Value())
	}
}

// TestLowerYouSubjectScrySequence proves a "you scry N" clause continuing a
// prior controller effect lowers as a controller scry, the form Overwhelmed
// Apprentice uses ("each opponent mills two cards. Then you scry 2.").
func TestLowerYouSubjectScrySequence(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Apprentice",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Each opponent mills two cards. Then you scry 2.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %d, want 2", len(mode.Sequence))
	}
	if _, ok := mode.Sequence[0].Primitive.(game.Mill); !ok {
		t.Fatalf("primitive[0] = %T, want game.Mill", mode.Sequence[0].Primitive)
	}
	scry, ok := mode.Sequence[1].Primitive.(game.Scry)
	if !ok {
		t.Fatalf("primitive[1] = %T, want game.Scry", mode.Sequence[1].Primitive)
	}
	if scry.Amount.Value() != 2 {
		t.Fatalf("scry.Amount = %d, want 2", scry.Amount.Value())
	}
	if scry.Player != game.ControllerReference() {
		t.Fatalf("scry.Player = %+v, want controller", scry.Player)
	}
}

// TestLowerPriorSubjectScryControllerSequence proves a fixed controller scry that
// trails a self-pump clause lowers as a controller scry rather than failing
// closed, the form Burning Prophet uses ("this creature gets +1/+0 until end of
// turn, then scry 1."): the verb's inherited self subject is ignored and the
// resolving controller scries.
func TestLowerPriorSubjectScryControllerSequence(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Prophet",
		Layout:     "normal",
		TypeLine:   "Creature",
		OracleText: "Whenever you cast a noncreature spell, this creature gets +1/+0 until end of turn, then scry 1.",
	})
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %d, want 2", len(mode.Sequence))
	}
	scry, ok := mode.Sequence[1].Primitive.(game.Scry)
	if !ok {
		t.Fatalf("primitive[1] = %T, want game.Scry", mode.Sequence[1].Primitive)
	}
	if scry.Amount.Value() != 1 || scry.Player != game.ControllerReference() {
		t.Fatalf("scry = %+v, want controller scries 1", scry)
	}
}

// TestLowerTargetSubjectScrySequence proves a fixed controller scry that trails a
// targeted pump lowers as a controller scry: the inherited target subject is
// ignored and the verb still acts on the resolving controller.
func TestLowerTargetSubjectScrySequence(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Target Scry",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Target creature gets +1/+0 until end of turn, then scry 1.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %d, want 2", len(mode.Sequence))
	}
	scry, ok := mode.Sequence[1].Primitive.(game.Scry)
	if !ok {
		t.Fatalf("primitive[1] = %T, want game.Scry", mode.Sequence[1].Primitive)
	}
	if scry.Player != game.ControllerReference() {
		t.Fatalf("scry.Player = %+v, want controller", scry.Player)
	}
}
