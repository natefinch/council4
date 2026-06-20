package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestManaAbilityTapRecordsTappedForManaProvenance verifies that activating a
// land's mana ability taps it with tapped-for-mana provenance on the emitted
// permanent-tapped event, so "is tapped for mana" triggers can distinguish it
// from an ordinary tap (Wild Growth and the mana-additional aura family).
func TestManaAbilityTapRecordsTappedForManaProvenance(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	land := addManaLandPermanent(g, game.Player1, "Forest", mana.G)

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(land.ObjectID, 0, nil, 0)) {
		t.Fatal("activating mana ability = false, want true")
	}
	assertEvent(t, g.Events, game.EventPermanentTapped, func(event game.Event) bool {
		return event.PermanentID == land.ObjectID && event.TappedForMana
	})
}

// TestNonManaTapDoesNotRecordTappedForMana verifies that tapping a permanent
// outside a mana ability leaves the provenance flag unset, so "is tapped for
// mana" triggers do not over-fire.
func TestNonManaTapDoesNotRecordTappedForMana(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	permanent := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Test Creature",
		Types: []types.Card{types.Creature},
	}})

	setPermanentTapped(g, permanent, true)

	assertEvent(t, g.Events, game.EventPermanentTapped, func(event game.Event) bool {
		return event.PermanentID == permanent.ObjectID && !event.TappedForMana
	})
}

// TestRequireTappedForManaPatternFiltersProvenance verifies the trigger pattern
// matcher only accepts a permanent-tapped event whose tap was for mana.
func TestRequireTappedForManaPatternFiltersProvenance(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Test Aura",
		Types: []types.Card{types.Enchantment},
	}})
	pattern := &game.TriggerPattern{
		Event:                game.EventPermanentTapped,
		RequireTappedForMana: true,
	}

	manaTap := game.Event{Kind: game.EventPermanentTapped, Controller: game.Player1, TappedForMana: true}
	if !triggerMatchesEvent(g, source, pattern, manaTap) {
		t.Fatal("tapped-for-mana event did not match RequireTappedForMana pattern")
	}
	plainTap := game.Event{Kind: game.EventPermanentTapped, Controller: game.Player1}
	if triggerMatchesEvent(g, source, pattern, plainTap) {
		t.Fatal("plain tap event matched RequireTappedForMana pattern")
	}
}
