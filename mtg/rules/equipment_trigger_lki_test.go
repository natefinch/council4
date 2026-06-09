package rules

import (
	"testing"

	cardb "github.com/natefinch/council4/mtg/cards/b"
	cardf "github.com/natefinch/council4/mtg/cards/f"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// addEquippedCreature creates a creature with explicit power and toughness,
// attaches the given equipment to it, and returns the creature.
func addEquippedCreature(g *game.Game, controller game.PlayerID, power, toughness int, equipment *game.Permanent) *game.Permanent {
	pt := game.PT{Value: power}
	creature := addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:      "Equipped Creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(pt),
		Toughness: opt.Val(game.PT{Value: toughness}),
	}})
	if !attachPermanent(g, equipment, creature) {
		panic("addEquippedCreature: attachPermanent returned false")
	}
	return creature
}

// TestFiendlashTriggerUsesLKIAfterLethalDamage verifies that when equipped
// creature receives lethal damage and dies before Fiendlash's triggered ability
// resolves, EventPermanentReference resolves via last-known information and the
// damage is dealt using the creature's captured power.
func TestFiendlashTriggerUsesLKIAfterLethalDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	// Power 4, toughness 3: dealt 3 damage → lethal; power stored in LKI.
	// Fiendlash adds +2/+0, so effective power = 6, toughness = 3.
	fiendlash := addCombatPermanent(g, game.Player1, cardf.Fiendlash)
	creature := addEquippedCreature(g, game.Player1, 4, 3, fiendlash)
	creatureObjectID := creature.ObjectID
	// Effective power includes Fiendlash's +2/+0 static ability.
	creatureEffectivePower := 6

	// Simulate the equipped creature receiving lethal damage.
	creature.MarkedDamage = 3
	g.AppendEvent(game.Event{
		Kind:            game.EventDamageDealt,
		PermanentID:     creature.ObjectID,
		DamageRecipient: game.DamageRecipientPermanent,
		Amount:          3,
	})

	// SBAs: creature dies with lethal damage. Equipment detaches; LKI snapshot
	// is recorded for the creature before it leaves the battlefield.
	engine.applyStateBasedActions(g)

	if fiendlash.AttachedTo.Exists {
		t.Fatalf("equipment still attached after creature death (AttachedTo = %v)", fiendlash.AttachedTo.Val)
	}
	if _, ok := permanentByObjectID(g, creatureObjectID); ok {
		t.Fatal("creature still on battlefield after lethal damage + SBAs")
	}

	// The LKI snapshot must record the creature's effective power (including Fiendlash's +2)
	// so EventPermanentReference can use it.
	snapshot, ok := lastKnownObject(g, creatureObjectID)
	if !ok {
		t.Fatal("no LKI snapshot for dead creature")
	}
	if !snapshot.Power.Exists || snapshot.Power.Val != creatureEffectivePower {
		t.Fatalf("LKI power = %v, want %d", snapshot.Power, creatureEffectivePower)
	}

	// Trigger detection: Fiendlash watches for damage dealt to its attached
	// permanent. It checks LKI to confirm the creature had Fiendlash attached.
	// Agent for Player1 selects Player2 (index 1 in the sorted player list).
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}},
	}
	log := TurnLog{}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &log) {
		t.Fatal("Fiendlash trigger was not put on stack after lethal damage event")
	}
	if g.Stack.Size() != 1 {
		t.Fatalf("stack size = %d, want 1 (Fiendlash trigger)", g.Stack.Size())
	}

	// Resolve: EventPermanentReference must retrieve creature's LKI power.
	engine.resolveTopOfStackWithChoices(g, agents, &log)

	// Fiendlash deals damage equal to the creature's effective power (6: 4 base + 2 from Fiendlash)
	// to Player2, using LKI captured at time of death.
	wantLife := 40 - creatureEffectivePower
	if got := g.Players[game.Player2].Life; got != wantLife {
		t.Fatalf("Player2 life = %d, want %d (creature effective power %d via LKI)", got, wantLife, creatureEffectivePower)
	}
}

// TestBlazingSunsteelTriggerUsesLKIAfterLethalDamage verifies that when
// equipped creature receives lethal damage and dies before Blazing Sunsteel's
// triggered ability resolves, the trigger still fires using EventPermanentReference
// for the damage source so the correct source identity is preserved.
func TestBlazingSunsteelTriggerUsesLKIAfterLethalDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	sunsteel := addCombatPermanent(g, game.Player1, cardb.BlazingSunsteel)
	creature := addEquippedCreature(g, game.Player1, 3, 3, sunsteel)
	creatureObjectID := creature.ObjectID
	eventDamageAmount := 5

	// The creature receives lethal damage from some source.
	creature.MarkedDamage = 3
	g.AppendEvent(game.Event{
		Kind:            game.EventDamageDealt,
		PermanentID:     creature.ObjectID,
		DamageRecipient: game.DamageRecipientPermanent,
		Amount:          eventDamageAmount,
	})

	// SBAs: creature dies; equipment detaches.
	engine.applyStateBasedActions(g)

	if sunsteel.AttachedTo.Exists {
		t.Fatal("equipment still attached after creature death")
	}
	if _, ok := permanentByObjectID(g, creatureObjectID); ok {
		t.Fatal("creature still on battlefield after lethal damage + SBAs")
	}

	// Agent for Player1 selects Player2 as target.
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}},
	}
	log := TurnLog{}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &log) {
		t.Fatal("Blazing Sunsteel trigger was not put on stack after lethal damage event")
	}

	engine.resolveTopOfStackWithChoices(g, agents, &log)

	// Blazing Sunsteel deals DynamicAmountEventDamage (5) to any target.
	wantLife := 40 - eventDamageAmount
	if got := g.Players[game.Player2].Life; got != wantLife {
		t.Fatalf("Player2 life = %d, want %d (event damage %d via trigger)", got, wantLife, eventDamageAmount)
	}
}
