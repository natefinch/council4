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
// when the attacking creature is goaded right now, and stays closed otherwise.
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

	// Goad the attacker; now the same attack event must satisfy the trigger.
	attacker.Goaded = map[game.PlayerID]game.GoadStatus{
		game.Player1: {CreatedTurn: 1, ExpiresFor: game.Player1},
	}
	if !triggerMatchesEvent(g, source, pattern, attacksEvent) {
		t.Fatal("goaded-subject trigger did not fire against a goaded attacker")
	}
}
