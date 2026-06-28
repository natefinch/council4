package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// optionalReferencedPlayerDraw asserts the lowered face carries exactly one
// triggered ability whose single instruction is an optional fixed-one Draw for
// the given player reference, with that same player as the OptionalActor (the
// player who decides whether to draw).
func optionalReferencedPlayerDraw(t *testing.T, face loweredFaceAbilities, want game.PlayerReference) {
	t.Helper()
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	modes := face.TriggeredAbilities[0].Content.Modes
	if len(modes) != 1 || len(modes[0].Sequence) != 1 {
		t.Fatalf("content modes/sequence = %#v, want one mode with one instruction", modes)
	}
	instr := modes[0].Sequence[0]
	draw, ok := instr.Primitive.(game.Draw)
	if !ok {
		t.Fatalf("primitive = %#v, want Draw", instr.Primitive)
	}
	if draw.Player != want {
		t.Fatalf("draw player = %#v, want %#v", draw.Player, want)
	}
	if draw.Amount.IsDynamic() || draw.Amount.Value() != 1 {
		t.Fatalf("draw amount = %#v, want fixed 1", draw.Amount)
	}
	if draw.PlayerGroup.Kind != game.PlayerGroupReferenceNone {
		t.Fatalf("draw player group = %#v, want none", draw.PlayerGroup)
	}
	if !instr.Optional {
		t.Fatal("instruction Optional = false, want true")
	}
	if !instr.OptionalActor.Exists || instr.OptionalActor.Val != want {
		t.Fatalf("instruction OptionalActor = %#v, want %#v", instr.OptionalActor, want)
	}
}

// TestLowerOptionalDefendingPlayerDraw proves "Whenever this creature attacks,
// defending player may draw a card." (Sibilant Spirit, Harbor Guardian) lowers
// to a single optional Draw for the defending player, who is also the player
// the engine asks whether to draw (OptionalActor).
func TestLowerOptionalDefendingPlayerDraw(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Sibilant Test",
		Layout:     "normal",
		TypeLine:   "Creature — Spirit",
		OracleText: "Flying\nWhenever this creature attacks, defending player may draw a card.",
	})
	optionalReferencedPlayerDraw(t, face, game.DefendingPlayerReference())
}

// TestLowerOptionalReferencedControllerDraw proves the combat-damage
// "its controller may draw a card." (Edric, Spymaster of Trest; Synapse Sliver)
// and the dies-trigger "that creature's controller may draw a card." (Fecundity)
// lower to a single optional Draw for the controller of the triggering event
// permanent, who is also the OptionalActor.
func TestLowerOptionalReferencedControllerDraw(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		typeLine   string
		oracleText string
	}{
		{
			name:       "combat damage its controller",
			typeLine:   "Legendary Creature — Elf Rogue",
			oracleText: "Whenever a creature deals combat damage to one of your opponents, its controller may draw a card.",
		},
		{
			name:       "dies that creature's controller",
			typeLine:   "Enchantment",
			oracleText: "Whenever a creature dies, that creature's controller may draw a card.",
		},
	}
	want := game.ObjectControllerReference(game.EventPermanentReference())
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Referenced Controller Draw " + test.name,
				Layout:     "normal",
				TypeLine:   test.typeLine,
				OracleText: test.oracleText,
			})
			optionalReferencedPlayerDraw(t, face, want)
		})
	}
}

// TestLowerOptionalReferencedPlayerDrawRejectsControllerSubject proves the
// recognizer does not claim the ordinary "you may draw a card" controller draw:
// that body keeps lowering through the established optional path, so the
// recognizer must fail closed for a controller-context draw rather than
// producing a spurious OptionalActor.
func TestLowerOptionalReferencedPlayerDrawRejectsControllerSubject(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Controller Draw Test",
		Layout:     "normal",
		TypeLine:   "Creature — Spirit",
		OracleText: "When this creature enters, you may draw a card.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	instr := face.TriggeredAbilities[0].Content.Modes[0].Sequence[0]
	draw, ok := instr.Primitive.(game.Draw)
	if !ok || draw.Player != game.ControllerReference() {
		t.Fatalf("draw = %#v, want controller draw", instr.Primitive)
	}
	if instr.OptionalActor.Exists {
		t.Fatalf("OptionalActor = %#v, want none for a controller draw", instr.OptionalActor)
	}
}
