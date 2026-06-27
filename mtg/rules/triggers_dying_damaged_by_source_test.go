package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
)

// dyingDamagedBySourcePattern mirrors the lowered shape of "Whenever a creature
// dealt damage by this creature this turn dies, ..." (Blood Cultist): a
// permanent-died trigger restricted to creatures the source permanent damaged
// earlier this turn.
func dyingDamagedBySourcePattern() *game.TriggerPattern {
	return &game.TriggerPattern{
		Event:                game.EventPermanentDied,
		DyingDamagedBySource: true,
		SubjectSelection:     game.Selection{RequiredTypes: []types.Card{types.Creature}},
	}
}

// addVictimCreature registers a creature permanent so a died event referencing
// it resolves to a creature-typed object for subject-selection matching.
func addVictimCreature(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:  "Victim",
		Types: []types.Card{types.Creature},
	}})
}

func emitSourceDamagedPermanent(g *game.Game, sourceObjectID, victimID id.ID) {
	emitEvent(g, game.Event{
		Kind:            game.EventDamageDealt,
		SourceObjectID:  sourceObjectID,
		PermanentID:     victimID,
		Amount:          1,
		DamageRecipient: game.DamageRecipientPermanent,
	})
}

func emitVictimDied(g *game.Game, victimID id.ID) {
	emitEvent(g, game.Event{
		Kind:        game.EventPermanentDied,
		PermanentID: victimID,
		CardTypes:   []types.Card{types.Creature},
	})
}

// TestDyingDamagedBySourceFiresOnlyForDamagedDeaths verifies the trigger fires
// when a creature the source damaged this turn dies, and ignores deaths of
// creatures the source never damaged.
func TestDyingDamagedBySourceFiresOnlyForDamagedDeaths(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Turn.ActivePlayer = game.Player1

	source := addTriggeredPermanent(g, game.Player1, dyingDamagedBySourcePattern(),
		[]game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	damaged := addVictimCreature(g, game.Player2)
	undamaged := addVictimCreature(g, game.Player2)

	emitSourceDamagedPermanent(g, source.ObjectID, damaged.ObjectID)

	// A creature the source never damaged dies: no trigger.
	emitVictimDied(g, undamaged.ObjectID)
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("trigger fired for a creature the source never damaged")
	}

	// The damaged creature dies: trigger fires.
	emitVictimDied(g, damaged.ObjectID)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("trigger did not fire for a creature the source damaged this turn")
	}
	if g.Stack.Size() != 1 {
		t.Fatalf("stack size = %d, want 1", g.Stack.Size())
	}
}

// TestDyingDamagedBySourceIgnoresOtherSourceDamage verifies damage dealt by a
// different permanent does not satisfy the "dealt damage by this creature"
// restriction.
func TestDyingDamagedBySourceIgnoresOtherSourceDamage(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Turn.ActivePlayer = game.Player1

	source := addTriggeredPermanent(g, game.Player1, dyingDamagedBySourcePattern(),
		[]game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	other := addVictimCreature(g, game.Player1)
	victim := addVictimCreature(g, game.Player2)

	// Damage comes from a different source permanent.
	emitSourceDamagedPermanent(g, other.ObjectID, victim.ObjectID)

	emitVictimDied(g, victim.ObjectID)
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("trigger fired for damage dealt by a different source")
	}
	_ = source
}
