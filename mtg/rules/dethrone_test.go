package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func dethroneCreature(name string) *game.CardDef {
	def := vanillaCreature(name, 2, 2)
	def.TriggeredAbilities = []game.TriggeredAbility{game.DethroneTriggeredBody}
	return def
}

// TestDethroneTriggerMatchesOnlyMostLifePlayer covers the Dethrone keyword
// (CR 702.103): the attacks trigger fires only when the attacked player has the
// most life or is tied for most among non-eliminated players.
func TestDethroneTriggerMatchesOnlyMostLifePlayer(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatPermanent(g, game.Player1, dethroneCreature("Treasonous Ogre"))
	pattern := &game.DethroneTriggeredBody.Trigger.Pattern

	attacks := func(target game.PlayerID) game.Event {
		return game.Event{
			Kind:         game.EventAttackerDeclared,
			Controller:   game.Player1,
			PermanentID:  attacker.ObjectID,
			AttackTarget: game.AttackTarget{Player: target},
		}
	}

	// All players start at 40 life, so any attacked player is tied for most.
	if !triggerMatchesEvent(g, attacker, pattern, attacks(game.Player2)) {
		t.Fatal("dethrone did not trigger against a player tied for most life")
	}

	// Raise a third player above the rest: attacking a lower-life player no
	// longer qualifies, while attacking the highest-life player does.
	g.Players[game.Player3].Life = 50
	if triggerMatchesEvent(g, attacker, pattern, attacks(game.Player2)) {
		t.Fatal("dethrone wrongly triggered against a player without the most life")
	}
	if !triggerMatchesEvent(g, attacker, pattern, attacks(game.Player3)) {
		t.Fatal("dethrone did not trigger against the player with the most life")
	}
}

// TestDethronePutsCounterOnAttacker confirms the canonical effect places a
// +1/+1 counter on the attacking creature itself (the ability source).
func TestDethronePutsCounterOnAttacker(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatPermanent(g, game.Player1, dethroneCreature("Treasonous Ogre"))

	add, ok := game.DethroneTriggeredBody.Content.Modes[0].Sequence[0].Primitive.(game.AddCounter)
	if !ok {
		t.Fatalf("dethrone content is not an AddCounter: %#v", game.DethroneTriggeredBody.Content.Modes[0].Sequence[0].Primitive)
	}
	stackObject := &game.StackObject{
		Controller: game.Player1,
		SourceID:   attacker.ObjectID,
	}
	resolver := newReferenceResolverWithSource(g, stackObject, attacker)
	resolved, ok := resolver.object(add.Object)
	if !ok || resolved.permanent == nil {
		t.Fatal("dethrone effect object reference did not resolve to a permanent")
	}
	if resolved.permanent.ObjectID != attacker.ObjectID {
		t.Fatalf("dethrone counter targeted %v, want attacker %v", resolved.permanent.ObjectID, attacker.ObjectID)
	}
}
