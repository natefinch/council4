package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// addSturdyCombatCreature adds a 1/10 creature that survives ordinary combat
// damage, so an end-of-combat delayed destroy is the only thing that can remove
// it and the assertions are not confounded by combat trades.
func addSturdyCombatCreature(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:      "Sturdy Combat Creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 10}),
	}})
}

// TestDelayedAtEndOfCombatCapturedEventPermanentDestroyed verifies the basilisk
// "deals combat damage to a creature, destroy that creature at end of combat"
// idiom: the creature named by the triggering event permanent is frozen at
// schedule time and destroyed at end of combat while the source survives.
func TestDelayedAtEndOfCombatCapturedEventPermanentDestroyed(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	basilisk := addSturdyCombatCreature(g, game.Player1)
	victim := addSturdyCombatCreature(g, game.Player2)
	if !scheduleDelayedTrigger(g, &game.StackObject{
		Kind:            game.StackTriggeredAbility,
		SourceID:        basilisk.ObjectID,
		SourceCardID:    basilisk.CardInstanceID,
		Controller:      game.Player1,
		HasTriggerEvent: true,
		TriggerEvent:    game.Event{PermanentID: victim.ObjectID},
	}, &game.DelayedTriggerDef{
		Timing:         game.DelayedAtEndOfCombat,
		CapturedObject: opt.Val(game.EventPermanentReference()),
		Content: game.Mode{
			Sequence: []game.Instruction{{Primitive: game.Destroy{Object: game.CapturedObjectReference()}}},
		}.Ability(),
	}) {
		t.Fatal("scheduleDelayedTrigger failed")
	}

	engine.runCombatPhase(g, allFirstLegalAgents(), &TurnLog{})

	if len(g.DelayedTriggers) != 0 {
		t.Fatalf("delayed triggers after end of combat = %d, want 0", len(g.DelayedTriggers))
	}
	if _, ok := permanentByObjectID(g, victim.ObjectID); ok {
		t.Fatal("captured creature remained on battlefield after end-of-combat destroy")
	}
	if !g.Players[game.Player2].Graveyard.Contains(victim.CardInstanceID) {
		t.Fatal("captured creature was not destroyed to its owner's graveyard")
	}
	if _, ok := permanentByObjectID(g, basilisk.ObjectID); !ok {
		t.Fatal("source permanent was incorrectly destroyed")
	}
}

// TestDelayedAtEndOfCombatCapturedEventRelatedPermanentDestroyed verifies the
// "blocks or becomes blocked by a creature, destroy that creature at end of
// combat" idiom, where "that creature" is the event's related permanent (the
// opposing combatant), frozen at schedule time and destroyed at end of combat.
func TestDelayedAtEndOfCombatCapturedEventRelatedPermanentDestroyed(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	basilisk := addSturdyCombatCreature(g, game.Player1)
	victim := addSturdyCombatCreature(g, game.Player2)
	if !scheduleDelayedTrigger(g, &game.StackObject{
		Kind:            game.StackTriggeredAbility,
		SourceID:        basilisk.ObjectID,
		SourceCardID:    basilisk.CardInstanceID,
		Controller:      game.Player1,
		HasTriggerEvent: true,
		TriggerEvent:    game.Event{RelatedPermanentID: victim.ObjectID},
	}, &game.DelayedTriggerDef{
		Timing:         game.DelayedAtEndOfCombat,
		CapturedObject: opt.Val(game.EventRelatedPermanentReference()),
		Content: game.Mode{
			Sequence: []game.Instruction{{Primitive: game.Destroy{Object: game.CapturedObjectReference()}}},
		}.Ability(),
	}) {
		t.Fatal("scheduleDelayedTrigger failed")
	}

	engine.runCombatPhase(g, allFirstLegalAgents(), &TurnLog{})

	if _, ok := permanentByObjectID(g, victim.ObjectID); ok {
		t.Fatal("captured related creature remained on battlefield after end-of-combat destroy")
	}
	if !g.Players[game.Player2].Graveyard.Contains(victim.CardInstanceID) {
		t.Fatal("captured related creature was not destroyed to its owner's graveyard")
	}
	if _, ok := permanentByObjectID(g, basilisk.ObjectID); !ok {
		t.Fatal("source permanent was incorrectly destroyed")
	}
}

// TestDelayedAtEndOfCombatCapturedObjectFailsClosedWhenGone verifies the delayed
// destroy is inert when the captured creature has already left the battlefield
// before end of combat: no other permanent is destroyed in its place.
func TestDelayedAtEndOfCombatCapturedObjectFailsClosedWhenGone(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	basilisk := addSturdyCombatCreature(g, game.Player1)
	victim := addSturdyCombatCreature(g, game.Player2)
	bystander := addSturdyCombatCreature(g, game.Player2)
	if !scheduleDelayedTrigger(g, &game.StackObject{
		Kind:            game.StackTriggeredAbility,
		SourceID:        basilisk.ObjectID,
		SourceCardID:    basilisk.CardInstanceID,
		Controller:      game.Player1,
		HasTriggerEvent: true,
		TriggerEvent:    game.Event{PermanentID: victim.ObjectID},
	}, &game.DelayedTriggerDef{
		Timing:         game.DelayedAtEndOfCombat,
		CapturedObject: opt.Val(game.EventPermanentReference()),
		Content: game.Mode{
			Sequence: []game.Instruction{{Primitive: game.Destroy{Object: game.CapturedObjectReference()}}},
		}.Ability(),
	}) {
		t.Fatal("scheduleDelayedTrigger failed")
	}
	if !movePermanentToZone(g, victim, zone.Graveyard) {
		t.Fatal("failed to move captured creature off the battlefield")
	}

	engine.runCombatPhase(g, allFirstLegalAgents(), &TurnLog{})

	if _, ok := permanentByObjectID(g, bystander.ObjectID); !ok {
		t.Fatal("delayed destroy of a departed captured creature hit another permanent")
	}
	if _, ok := permanentByObjectID(g, basilisk.ObjectID); !ok {
		t.Fatal("delayed destroy of a departed captured creature hit the source")
	}
}
