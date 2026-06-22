package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// TestEventPermanentPowerDamageReadsEnteringCreature proves the Terror of the
// Peaks shape: "Whenever another creature you control enters, ~ deals damage
// equal to that creature's power to any target." The damage amount reads the
// power of the permanent named by the triggering enters event (the entering
// creature), while the damage source remains the ability's own source. The
// chosen target therefore takes damage equal to the entering creature's power,
// not the source's power.
func TestEventPermanentPowerDamageReadsEnteringCreature(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatCreaturePermanentWithPower(g, game.Player1, 7)
	entering := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	victim := addCombatCreaturePermanentWithPower(g, game.Player2, 10)

	obj := &game.StackObject{
		ID:              g.IDGen.Next(),
		Controller:      game.Player1,
		SourceID:        source.ObjectID,
		HasTriggerEvent: true,
		TriggerEvent: game.Event{
			Kind:        game.EventPermanentEnteredBattlefield,
			PermanentID: entering.ObjectID,
		},
		Targets: []game.Target{game.PermanentTarget(victim.ObjectID)},
	}

	resolveInstruction(engine, g, obj, game.Damage{
		Amount: game.Dynamic(game.DynamicAmount{
			Kind:       game.DynamicAmountObjectPower,
			Multiplier: 1,
			Object:     game.EventPermanentReference(),
		}),
		Recipient:    game.AnyTargetDamageRecipient(0),
		DamageSource: opt.Val(game.SourcePermanentReference()),
	}, &TurnLog{})

	if victim.MarkedDamage != 3 {
		t.Fatalf("victim marked damage = %d, want 3 (entering creature power, not source power 7)", victim.MarkedDamage)
	}
	if entering.MarkedDamage != 0 {
		t.Fatalf("entering creature marked damage = %d, want 0", entering.MarkedDamage)
	}
}
