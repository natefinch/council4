package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// vengefulAncestorSource builds a battlefield permanent standing in for Vengeful
// Ancestor so its trigger patterns can be matched against events.
func vengefulAncestorSource(g *game.Game) *game.Permanent {
	return addCombatPermanent(g, game.Player1, vanillaCreature("Vengeful Ancestor", 3, 4))
}

// TestVengefulAncestorUnionTriggerFiresOnEntersAndAttacks proves the
// "enters or attacks" union trigger fires on BOTH constituent events: the same
// pattern (Event enters, UnionEvent attacks, TriggerSourceSelf) matches an
// entered-battlefield event for the source and an attacker-declared event for
// the source.
func TestVengefulAncestorUnionTriggerFiresOnEntersAndAttacks(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := vengefulAncestorSource(g)
	pattern := &game.TriggerPattern{
		Event:      game.EventPermanentEnteredBattlefield,
		Source:     game.TriggerSourceSelf,
		UnionEvent: game.EventAttackerDeclared,
	}

	entersEvent := game.Event{
		Kind:        game.EventPermanentEnteredBattlefield,
		Controller:  game.Player1,
		PermanentID: source.ObjectID,
	}
	if !triggerMatchesEvent(g, source, pattern, entersEvent) {
		t.Fatal("union trigger did not fire on the enters event")
	}

	attacksEvent := game.Event{
		Kind:         game.EventAttackerDeclared,
		Controller:   game.Player1,
		PermanentID:  source.ObjectID,
		AttackTarget: game.AttackTarget{Player: game.Player2},
	}
	if !triggerMatchesEvent(g, source, pattern, attacksEvent) {
		t.Fatal("union trigger did not fire on the attacks event")
	}
}

// TestVengefulAncestorGoadedSubjectTriggerRequiresGoad covers the goaded
// trigger-subject qualifier: "Whenever a goaded creature attacks" fires only
// when the attack event records that the creature was goaded.
func TestVengefulAncestorGoadedSubjectTriggerRequiresGoad(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := vengefulAncestorSource(g)
	attacker := addCombatPermanent(g, game.Player2, vanillaCreature("Goaded Attacker", 2, 2))
	pattern := &game.TriggerPattern{
		Event: game.EventAttackerDeclared,
		SubjectSelection: game.Selection{
			RequiredTypes: []types.Card{types.Creature},
			MatchGoaded:   true,
		},
	}
	attacksEvent := game.Event{
		Kind:         game.EventAttackerDeclared,
		Controller:   game.Player2,
		PermanentID:  attacker.ObjectID,
		AttackTarget: game.AttackTarget{Player: game.Player1},
	}

	// The attacker is not goaded yet, so the goaded-subject trigger must not fire.
	if triggerMatchesEvent(g, source, pattern, attacksEvent) {
		t.Fatal("goaded-subject trigger fired against a creature that is not goaded")
	}

	// An authoritative false snapshot remains false even if the permanent is
	// goaded later.
	attacker.Goaded = map[game.PlayerID]game.GoadStatus{
		game.Player1: {CreatedTurn: 1, ExpiresFor: game.Player1},
	}
	attacksEvent.SubjectGoadedKnown = true
	if triggerMatchesEvent(g, source, pattern, attacksEvent) {
		t.Fatal("goaded-subject trigger used goad added after the attack event")
	}

	// An authoritative true snapshot remains true if the goad later disappears.
	attacksEvent.SubjectGoaded = true
	attacker.Goaded = nil
	if !triggerMatchesEvent(g, source, pattern, attacksEvent) {
		t.Fatal("goaded-subject trigger did not fire against a goaded attacker")
	}
}

func TestRelatedSubjectGoadDoesNotReusePrimaryEventSnapshot(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	primary := addCombatPermanent(g, game.Player2, vanillaCreature("Primary", 2, 2))
	related := addCombatPermanent(g, game.Player3, vanillaCreature("Related", 2, 2))
	event := game.Event{
		Kind:               game.EventBlockerDeclared,
		Controller:         game.Player2,
		PermanentID:        primary.ObjectID,
		RelatedPermanentID: related.ObjectID,
		SubjectGoaded:      true,
		SubjectGoadedKnown: true,
	}
	selection := game.Selection{
		RequiredTypes: []types.Card{types.Creature},
		MatchGoaded:   true,
	}

	if triggerSelectionMatches(g, game.Player1, event, related.ObjectID, &selection, 0) {
		t.Fatal("related subject reused the primary event subject's goad snapshot")
	}
	related.Goaded = map[game.PlayerID]game.GoadStatus{
		game.Player1: {RestOfGame: true},
	}
	if !triggerSelectionMatches(g, game.Player1, event, related.ObjectID, &selection, 0) {
		t.Fatal("related subject did not use its own live goad status")
	}
}
