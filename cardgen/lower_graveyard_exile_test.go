package cardgen

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestLowerTargetedGraveyardExile(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Wretch",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Exile target card from a graveyard.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %#v, want one", mode.Targets)
	}
	target := mode.Targets[0]
	if target.MinTargets != 1 || target.MaxTargets != 1 ||
		target.Allow != game.TargetAllowCard || target.TargetZone != zone.Graveyard ||
		!target.Selection.Val.Empty() {
		t.Fatalf("target = %#v", target)
	}
	move, ok := mode.Sequence[0].Primitive.(game.MoveCard)
	if !ok {
		t.Fatalf("primitive = %T, want game.MoveCard", mode.Sequence[0].Primitive)
	}
	if move.Card.Kind != game.CardReferenceTarget || move.Card.TargetIndex != 0 ||
		move.FromZone != zone.Graveyard || move.Destination != zone.Exile {
		t.Fatalf("move = %#v", move)
	}
}

func TestLowerTargetedGraveyardExileTypedFromOpponent(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Mummy",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Exile target creature card from an opponent's graveyard.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	target := mode.Targets[0]
	if target.Allow != game.TargetAllowCard || target.TargetZone != zone.Graveyard {
		t.Fatalf("target = %#v", target)
	}
	selection := target.Selection.Val
	if !slices.Equal(selection.RequiredTypes, []types.Card{types.Creature}) ||
		selection.Controller != game.ControllerOpponent {
		t.Fatalf("selection = %#v", selection)
	}
	move, ok := mode.Sequence[0].Primitive.(game.MoveCard)
	if !ok || move.FromZone != zone.Graveyard || move.Destination != zone.Exile {
		t.Fatalf("move = %#v", mode.Sequence[0].Primitive)
	}
}

func TestLowerTargetedGraveyardExileUpToOne(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Gryff",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Exile up to one target card from a graveyard.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	target := mode.Targets[0]
	if target.MinTargets != 0 || target.MaxTargets != 1 ||
		target.Allow != game.TargetAllowCard || target.TargetZone != zone.Graveyard {
		t.Fatalf("target = %#v", target)
	}
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence length = %d, want 1", len(mode.Sequence))
	}
	move, ok := mode.Sequence[0].Primitive.(game.MoveCard)
	if !ok || move.Card.TargetIndex != 0 || move.Destination != zone.Exile {
		t.Fatalf("move = %#v", mode.Sequence[0].Primitive)
	}
}

// TestLowerTargetedGraveyardExileFailsClosedSingleGraveyard documents that the
// "from a single graveyard" constraint (all targets share one graveyard) is a
// distinct targeting restriction the canonical owner-suffix reconstruction does
// not render, so it stays unsupported rather than lowering to a wrong predicate.
func TestLowerTargetedGraveyardExileFailsClosedSingleGraveyard(t *testing.T) {
	t.Parallel()
	_, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Decay",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Exile up to three target cards from a single graveyard.",
	})
	if len(diagnostics) == 0 {
		t.Fatal("expected unsupported diagnostic for single-graveyard exile")
	}
}

// TestLowerTargetedGraveyardExileFailsClosedPlayerGraveyard documents that
// exiling an entire player's graveyard ("Exile target player's graveyard.") is
// a player-targeted zone wipe, not a card target, so it stays unsupported.
func TestLowerTargetedGraveyardExileFailsClosedPlayerGraveyard(t *testing.T) {
	t.Parallel()
	_, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Bog",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Exile target player's graveyard.",
	})
	if len(diagnostics) == 0 {
		t.Fatal("expected unsupported diagnostic for player-graveyard exile")
	}
}
