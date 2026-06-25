package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// firstSpellSequence returns the instruction sequence of the first mode of a
// lowered spell ability so a scope test can inspect the emitted primitives.
func firstSpellSequence(t *testing.T, face loweredFaceAbilities) []game.Instruction {
	t.Helper()
	if !face.SpellAbility.Exists || len(face.SpellAbility.Val.Modes) == 0 {
		t.Fatal("expected a spell ability with at least one mode")
	}
	return face.SpellAbility.Val.Modes[0].Sequence
}

// TestLowerGraveyardReturnScopesShareDestination proves that the graveyard-return
// scopes still emit their distinguishing primitive routing after the per-card
// destination reconstruction (MoveCard-to-hand, PutOnBattlefield, and the
// MassReturnFromGraveyard / ChooseFromZone group forms) was converged onto the
// shared graveyardReturnInstruction helper and plainGraveyardReturn precondition.
// A regression in the shared helper would drop or mis-shape one of these without
// necessarily slipping past the corpus gate.
func TestLowerGraveyardReturnScopesShareDestination(t *testing.T) {
	t.Parallel()

	t.Run("targeted reanimate to battlefield", func(t *testing.T) {
		t.Parallel()
		face := lowerSingleFace(t, &ScryfallCard{
			Name:       "Test GY Reanimate",
			Layout:     "normal",
			TypeLine:   "Sorcery",
			OracleText: "Return target creature card from your graveyard to the battlefield.",
		})
		mode := face.SpellAbility.Val.Modes[0]
		if len(mode.Targets) != 1 {
			t.Fatalf("targets = %d, want 1", len(mode.Targets))
		}
		seq := firstSpellSequence(t, face)
		if len(seq) != 1 {
			t.Fatalf("sequence length = %d, want 1", len(seq))
		}
		put, ok := seq[0].Primitive.(game.PutOnBattlefield)
		if !ok {
			t.Fatalf("primitive = %T, want game.PutOnBattlefield", seq[0].Primitive)
		}
		want := game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 0})
		if put.Source != want {
			t.Fatalf("put.Source = %#v, want %#v", put.Source, want)
		}
	})

	t.Run("targeted return to hand", func(t *testing.T) {
		t.Parallel()
		face := lowerSingleFace(t, &ScryfallCard{
			Name:       "Test GY To Hand",
			Layout:     "normal",
			TypeLine:   "Sorcery",
			OracleText: "Return target creature card from your graveyard to your hand.",
		})
		mode := face.SpellAbility.Val.Modes[0]
		if len(mode.Targets) != 1 {
			t.Fatalf("targets = %d, want 1", len(mode.Targets))
		}
		seq := firstSpellSequence(t, face)
		if len(seq) != 1 {
			t.Fatalf("sequence length = %d, want 1", len(seq))
		}
		move, ok := seq[0].Primitive.(game.MoveCard)
		if !ok {
			t.Fatalf("primitive = %T, want game.MoveCard", seq[0].Primitive)
		}
		if move.FromZone != zone.Graveyard || move.Destination != zone.Hand {
			t.Fatalf("move zones = %v->%v, want graveyard->hand", move.FromZone, move.Destination)
		}
		wantCard := game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 0}
		if move.Card != wantCard {
			t.Fatalf("move.Card = %#v, want %#v", move.Card, wantCard)
		}
	})

	t.Run("chosen at resolution to hand", func(t *testing.T) {
		t.Parallel()
		face := lowerSingleFace(t, &ScryfallCard{
			Name:       "Test GY Choice",
			Layout:     "normal",
			TypeLine:   "Sorcery",
			OracleText: "Return a creature card from your graveyard to your hand.",
		})
		mode := face.SpellAbility.Val.Modes[0]
		if len(mode.Targets) != 0 {
			t.Fatalf("targets = %d, want 0", len(mode.Targets))
		}
		seq := firstSpellSequence(t, face)
		if len(seq) != 1 {
			t.Fatalf("sequence length = %d, want 1", len(seq))
		}
		if _, ok := seq[0].Primitive.(game.ChooseFromZone); !ok {
			t.Fatalf("primitive = %T, want game.ChooseFromZone", seq[0].Primitive)
		}
	})

	t.Run("mass return to battlefield", func(t *testing.T) {
		t.Parallel()
		face := lowerSingleFace(t, &ScryfallCard{
			Name:       "Test GY Mass",
			Layout:     "normal",
			TypeLine:   "Sorcery",
			OracleText: "Return all creature cards from your graveyard to the battlefield.",
		})
		mode := face.SpellAbility.Val.Modes[0]
		if len(mode.Targets) != 0 {
			t.Fatalf("targets = %d, want 0", len(mode.Targets))
		}
		seq := firstSpellSequence(t, face)
		if len(seq) != 1 {
			t.Fatalf("sequence length = %d, want 1", len(seq))
		}
		mass, ok := seq[0].Primitive.(game.MassReturnFromGraveyard)
		if !ok {
			t.Fatalf("primitive = %T, want game.MassReturnFromGraveyard", seq[0].Primitive)
		}
		if mass.Destination != zone.Battlefield {
			t.Fatalf("mass.Destination = %v, want battlefield", mass.Destination)
		}
		if mass.ControlledByOwner {
			t.Fatal("mass.ControlledByOwner = true, want false for a your-graveyard mass return")
		}
	})
}
