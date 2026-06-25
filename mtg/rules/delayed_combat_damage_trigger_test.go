package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/opt"
)

// scheduleCapturedCombatDamageTrigger schedules a delayed "whenever that
// creature deals combat damage to a player this turn, you draw a card" trigger
// bound to the permanent identified by boundObjectID, mirroring the runtime
// shape lowering produces: a prior clause publishes the pumped creature under
// linkKey, and the delayed trigger captures it through a linked-object
// reference.
func scheduleCapturedCombatDamageTrigger(g *game.Game, source *game.StackObject, linkKey string, boundObjectID id.ID) {
	rememberLinkedObject(g, linkedObjectSourceKey(g, source, linkKey), game.LinkedObjectRef{ObjectID: boundObjectID})
	def := &game.DelayedTriggerDef{
		EventPattern: opt.Val(game.TriggerPattern{
			Event:                game.EventDamageDealt,
			RequireCombatDamage:  true,
			DamageRecipient:      game.DamageRecipientPlayer,
			DamageSourceCaptured: true,
		}),
		Window:             game.DelayedWindowThisTurn,
		DamageSourceObject: opt.Val(game.LinkedObjectReference(linkKey)),
		Content: game.Mode{Sequence: []game.Instruction{{
			Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()},
		}}}.Ability(),
	}
	if !scheduleDelayedTrigger(g, source, def) {
		panic("scheduleDelayedTrigger returned false")
	}
}

// TestCapturedCombatDamageTriggerFiresOnlyForBoundCreature verifies an
// object-identity-bound delayed combat-damage trigger ("... target creature ...
// Whenever that creature deals combat damage to a player this turn, you draw a
// card") fires when the captured creature deals combat damage to a player,
// ignores combat damage from a different creature, and ignores noncombat damage
// from the captured creature.
func TestCapturedCombatDamageTriggerFiresOnlyForBoundCreature(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Turn.ActivePlayer = game.Player1
	boundObjectID := g.IDGen.Next()
	otherObjectID := g.IDGen.Next()
	source := &game.StackObject{ID: g.IDGen.Next(), SourceID: g.IDGen.Next(), Controller: game.Player1}
	scheduleCapturedCombatDamageTrigger(g, source, "delayed-target-1", boundObjectID)

	emitEvent(g, game.Event{
		Kind:            game.EventDamageDealt,
		SourceObjectID:  otherObjectID,
		CombatDamage:    true,
		DamageRecipient: game.DamageRecipientPlayer,
		Player:          game.Player2,
	})
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("combat damage from a different creature fired the captured trigger")
	}

	emitEvent(g, game.Event{
		Kind:            game.EventDamageDealt,
		SourceObjectID:  boundObjectID,
		DamageRecipient: game.DamageRecipientPlayer,
		Player:          game.Player2,
	})
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("noncombat damage from the captured creature fired the captured trigger")
	}

	emitEvent(g, game.Event{
		Kind:            game.EventDamageDealt,
		SourceObjectID:  boundObjectID,
		CombatDamage:    true,
		DamageRecipient: game.DamageRecipientPlayer,
		Player:          game.Player2,
	})
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("captured creature's combat damage to a player did not fire the trigger")
	}
	if g.Stack.Size() != 1 {
		t.Fatalf("stack size = %d, want 1", g.Stack.Size())
	}
	g.Stack.Pop()

	expireEventDelayedTriggers(g)
	emitEvent(g, game.Event{
		Kind:            game.EventDamageDealt,
		SourceObjectID:  boundObjectID,
		CombatDamage:    true,
		DamageRecipient: game.DamageRecipientPlayer,
		Player:          game.Player2,
	})
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("captured combat-damage trigger fired after its this-turn window ended")
	}
}

// TestCapturedCombatDamageTriggerNeverFiresWhenCreatureGone verifies the bound
// trigger never fires when the captured permanent's identity was not recorded,
// so a vanished creature's rider stays dormant rather than firing on unrelated
// combat damage.
func TestCapturedCombatDamageTriggerNeverFiresWhenCreatureGone(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Turn.ActivePlayer = game.Player1
	source := &game.StackObject{ID: g.IDGen.Next(), SourceID: g.IDGen.Next(), Controller: game.Player1}
	def := &game.DelayedTriggerDef{
		EventPattern: opt.Val(game.TriggerPattern{
			Event:                game.EventDamageDealt,
			RequireCombatDamage:  true,
			DamageRecipient:      game.DamageRecipientPlayer,
			DamageSourceCaptured: true,
		}),
		Window:             game.DelayedWindowThisTurn,
		DamageSourceObject: opt.Val(game.LinkedObjectReference("delayed-target-1")),
		Content: game.Mode{Sequence: []game.Instruction{{
			Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()},
		}}}.Ability(),
	}
	if !scheduleDelayedTrigger(g, source, def) {
		t.Fatal("scheduleDelayedTrigger returned false")
	}

	emitEvent(g, game.Event{
		Kind:            game.EventDamageDealt,
		SourceObjectID:  g.IDGen.Next(),
		CombatDamage:    true,
		DamageRecipient: game.DamageRecipientPlayer,
		Player:          game.Player2,
	})
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("unbound captured trigger fired on unrelated combat damage")
	}
}
