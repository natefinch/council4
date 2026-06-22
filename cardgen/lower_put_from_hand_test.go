package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestLowerGrowthSpiralPutsLandFromHand(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Growth Spiral",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Draw a card. You may put a land card from your hand onto the battlefield.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("Growth Spiral did not lower to a spell ability")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want draw then put-from-hand", mode.Sequence)
	}
	if _, ok := mode.Sequence[0].Primitive.(game.Draw); !ok {
		t.Fatalf("first primitive = %T, want Draw", mode.Sequence[0].Primitive)
	}
	put, ok := mode.Sequence[1].Primitive.(game.PutFromHand)
	if !ok {
		t.Fatalf("second primitive = %T, want PutFromHand", mode.Sequence[1].Primitive)
	}
	if !mode.Sequence[1].Optional {
		t.Fatal("put-from-hand step is not optional (the \"you may\" wrapper was lost)")
	}
	if put.Player.Kind() != game.PlayerReferenceController || put.Amount.Value() != 1 {
		t.Fatalf("put = %#v, want controller put one", put)
	}
	if len(put.Selection.RequiredTypes) != 1 || put.Selection.RequiredTypes[0] != types.Land {
		t.Fatalf("selection = %#v, want land card", put.Selection)
	}
}

func TestLowerMandatoryPutLandFromHand(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Ramp",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Put a land card from your hand onto the battlefield.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence = %#v, want single put-from-hand", mode.Sequence)
	}
	if _, ok := mode.Sequence[0].Primitive.(game.PutFromHand); !ok {
		t.Fatalf("primitive = %T, want PutFromHand", mode.Sequence[0].Primitive)
	}
	if mode.Sequence[0].Optional {
		t.Fatal("mandatory put-from-hand was marked optional")
	}
}

func TestPutFromHandFailsClosedForLibrarySource(t *testing.T) {
	t.Parallel()
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Bad Ramp",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Put a land card from your library onto the battlefield.",
	})
}

func TestLowerMandatoryPutLandFromHandTapped(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Tapped Ramp",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Put a land card from your hand onto the battlefield tapped.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence = %#v, want single put-from-hand", mode.Sequence)
	}
	put, ok := mode.Sequence[0].Primitive.(game.PutFromHand)
	if !ok {
		t.Fatalf("primitive = %T, want PutFromHand", mode.Sequence[0].Primitive)
	}
	if !put.EntersTapped {
		t.Fatal("put-from-hand did not carry the \"tapped\" entry rider")
	}
	if mode.Sequence[0].Optional {
		t.Fatal("mandatory put-from-hand was marked optional")
	}
	if len(put.Selection.RequiredTypes) != 1 || put.Selection.RequiredTypes[0] != types.Land {
		t.Fatalf("selection = %#v, want land card", put.Selection)
	}
}

func TestLowerOptionalPutLandFromHandTapped(t *testing.T) {
	t.Parallel()
	// Horizon of Progress: an activated ability whose sole resolving effect is an
	// optional put-from-hand onto the battlefield tapped.
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Horizon",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "{3}, {T}: You may put a land card from your hand onto the battlefield tapped.",
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
	}
	mode := face.ActivatedAbilities[0].Content.Modes[0]
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence = %#v, want single put-from-hand", mode.Sequence)
	}
	put, ok := mode.Sequence[0].Primitive.(game.PutFromHand)
	if !ok {
		t.Fatalf("primitive = %T, want PutFromHand", mode.Sequence[0].Primitive)
	}
	if !put.EntersTapped {
		t.Fatal("put-from-hand did not carry the \"tapped\" entry rider")
	}
	if !mode.Sequence[0].Optional {
		t.Fatal("put-from-hand step is not optional (the \"you may\" wrapper was lost)")
	}
}
