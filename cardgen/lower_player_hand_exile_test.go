package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// TestLowerTargetPlayerHandExileTrigger proves the "target opponent exiles a
// card from their hand" trigger (Unscrupulous Agent, Skullcap Snail) lowers to
// one ChooseFromZone whose chooser is the targeted player, exiling a fixed one
// card from that player's own hand. The redundant "their" possessive is already
// expressed by scoping the choice to the target, so it does not block lowering.
func TestLowerTargetPlayerHandExileTrigger(t *testing.T) {
	t.Parallel()
	power := "1"
	toughness := "1"
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Hand Exiler",
		Layout:     "normal",
		TypeLine:   "Creature — Elf Detective",
		ManaCost:   "{1}{B}",
		OracleText: "When this creature enters, target opponent exiles a card from their hand.",
		Power:      &power,
		Toughness:  &toughness,
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	if len(mode.Targets) != 1 || mode.Targets[0].Allow != game.TargetAllowPlayer {
		t.Fatalf("targets = %+v, want one player target", mode.Targets)
	}
	choose, ok := mode.Sequence[0].Primitive.(game.ChooseFromZone)
	if !ok {
		t.Fatalf("primitive = %T, want game.ChooseFromZone", mode.Sequence[0].Primitive)
	}
	if choose.Player != game.TargetPlayerReference(0) {
		t.Fatalf("chooser = %v, want TargetPlayerReference(0)", choose.Player)
	}
	if choose.SourceZone != zone.Hand {
		t.Fatalf("source zone = %v, want hand", choose.SourceZone)
	}
	if choose.Destination.Zone != zone.Exile {
		t.Fatalf("destination = %v, want exile", choose.Destination.Zone)
	}
	if choose.Quantity != game.Fixed(1) {
		t.Fatalf("quantity = %v, want Fixed(1)", choose.Quantity)
	}
}

// TestLowerTargetPlayerHandExileActivatedTwo proves the activated form exiling a
// fixed two cards (Vessel of Malignity) carries the plural amount through to the
// ChooseFromZone quantity, confirming the lowerer reads the parsed fixed amount
// rather than assuming one card.
func TestLowerTargetPlayerHandExileActivatedTwo(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Vessel",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		ManaCost:   "{B}",
		OracleText: "{1}{B}, Sacrifice this enchantment: Target opponent exiles two cards from their hand. Activate only as a sorcery.",
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
	}
	mode := face.ActivatedAbilities[0].Content.Modes[0]
	choose, ok := mode.Sequence[0].Primitive.(game.ChooseFromZone)
	if !ok {
		t.Fatalf("primitive = %T, want game.ChooseFromZone", mode.Sequence[0].Primitive)
	}
	if choose.Player != game.TargetPlayerReference(0) {
		t.Fatalf("chooser = %v, want TargetPlayerReference(0)", choose.Player)
	}
	if choose.Quantity != game.Fixed(2) {
		t.Fatalf("quantity = %v, want Fixed(2)", choose.Quantity)
	}
}

// TestLowerTargetPlayerHandExileFailsClosedNonHand proves the lowerer does not
// fire for a graveyard-zone exile, whose parsed selector is not hand-scoped, so
// that family keeps failing closed rather than being mislowered as a hand exile.
func TestLowerTargetPlayerHandExileFailsClosedNonHand(t *testing.T) {
	t.Parallel()
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Test Graveyard Exiler",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{B}",
		OracleText: "Target opponent exiles two cards from their graveyard.",
	})
}
