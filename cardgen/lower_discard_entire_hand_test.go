package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestLowerDiscardEntireHandController(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "One with Nothing",
		Layout:     "normal",
		ManaCost:   "{B}",
		TypeLine:   "Instant",
		OracleText: "Discard your hand.",
	})
	sequence := face.SpellAbility.Val.Modes[0].Sequence
	if len(sequence) != 1 {
		t.Fatalf("sequence = %#v, want single discard", sequence)
	}
	discard, ok := sequence[0].Primitive.(game.Discard)
	if !ok || !discard.EntireHand || discard.Player != game.ControllerReference() {
		t.Fatalf("instruction 0 = %#v, want controller discard entire hand", sequence[0])
	}
	if discard.Amount.IsDynamic() || discard.Amount.Value() != 0 {
		t.Fatalf("discard amount = %#v, want zero", discard.Amount)
	}
}

func TestLowerDiscardEntireHandEachPlayer(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Sire of Insanity",
		Layout:     "normal",
		ManaCost:   "{4}{B}{B}",
		TypeLine:   "Creature — Demon",
		OracleText: "At the beginning of each end step, each player discards their hand.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %#v, want one", face.TriggeredAbilities)
	}
	sequence := face.TriggeredAbilities[0].Content.Modes[0].Sequence
	discard, ok := sequence[0].Primitive.(game.Discard)
	if !ok || !discard.EntireHand || discard.PlayerGroup.Kind != game.PlayerGroupReferenceAllPlayers {
		t.Fatalf("instruction 0 = %#v, want all-players discard entire hand", sequence[0])
	}
}

func TestLowerDiscardEntireHandTargetPlayer(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Wit's End",
		Layout:     "normal",
		ManaCost:   "{5}{B}{B}",
		TypeLine:   "Sorcery",
		OracleText: "Target player discards their hand.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %#v, want one player target", mode.Targets)
	}
	discard, ok := mode.Sequence[0].Primitive.(game.Discard)
	if !ok || !discard.EntireHand || discard.Player != game.TargetPlayerReference(0) {
		t.Fatalf("instruction 0 = %#v, want target-player discard entire hand", mode.Sequence[0])
	}
}

// TestLowerWindfallDiscardDrawGreatestThisWay covers the Windfall pattern
// "Each player discards their hand, then draws cards equal to the greatest
// number of cards a player discarded this way.": a group discard publishing the
// greatest per-player discard count followed by a group draw that reads it.
func TestLowerWindfallDiscardDrawGreatestThisWay(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Windfall",
		Layout:     "normal",
		ManaCost:   "{2}{U}",
		TypeLine:   "Sorcery",
		OracleText: "Each player discards their hand, then draws cards equal to the greatest number of cards a player discarded this way.",
	})
	sequence := face.SpellAbility.Val.Modes[0].Sequence
	if len(sequence) != 2 {
		t.Fatalf("sequence = %#v, want discard then draw", sequence)
	}
	discard, ok := sequence[0].Primitive.(game.Discard)
	if !ok || !discard.EntireHand || discard.PlayerGroup.Kind != game.PlayerGroupReferenceAllPlayers {
		t.Fatalf("instruction 0 = %#v, want all-players discard entire hand", sequence[0])
	}
	if sequence[0].PublishResult == "" {
		t.Fatalf("discard instruction = %#v, want a published result key", sequence[0])
	}
	draw, ok := sequence[1].Primitive.(game.Draw)
	if !ok || draw.PlayerGroup.Kind != game.PlayerGroupReferenceAllPlayers {
		t.Fatalf("instruction 1 = %#v, want all-players draw", sequence[1])
	}
	if !draw.Amount.IsDynamic() {
		t.Fatalf("draw amount = %#v, want dynamic", draw.Amount)
	}
	dynamic := draw.Amount.DynamicAmount().Val
	if dynamic.Kind != game.DynamicAmountPreviousEffectResult ||
		dynamic.ResultKey != sequence[0].PublishResult {
		t.Fatalf("draw amount = %#v, want previous-effect result keyed to the discard", dynamic)
	}
}
