package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// countControlledCreatures builds the dynamic +1/+1-per-creature amount used by
// self and event-permanent pumps ("where X is the number of creatures you
// control" / "for each creature you control").
func countControlledCreatures() game.DynamicAmount {
	return game.DynamicAmount{
		Kind:       game.DynamicAmountCountSelector,
		Multiplier: 1,
		Group: game.BattlefieldGroup(game.Selection{
			RequiredTypes: []types.Card{types.Creature},
			Controller:    game.ControllerYou,
		}),
	}
}

// TestSourceDynamicSelfPumpScalesWithBoardAndExpires proves that a self-pump
// ("This creature gets +X/+X until end of turn, where X is the number of
// creatures you control.") reads the live board at resolution, applies the
// computed +X/+X to the source permanent, and reverts during cleanup.
func TestSourceDynamicSelfPumpScalesWithBoardAndExpires(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	addCombatCreaturePermanentWithPower(g, game.Player2, 2)

	obj := &game.StackObject{
		Kind:         game.StackActivatedAbility,
		SourceID:     source.ObjectID,
		SourceCardID: source.CardInstanceID,
		Controller:   game.Player1,
	}
	count := countControlledCreatures()
	resolveInstruction(engine, g, obj, game.ModifyPT{
		Object:         game.SourcePermanentReference(),
		PowerDelta:     game.Dynamic(count),
		ToughnessDelta: game.Dynamic(count),
		Duration:       game.DurationUntilEndOfTurn,
	}, &TurnLog{})

	if got := effectivePower(g, source); got != 5 {
		t.Fatalf("effective power = %d, want 5 (2 base + 3 controlled creatures)", got)
	}
	if got, _ := effectiveToughness(g, source); got != 5 {
		t.Fatalf("effective toughness = %d, want 5 (2 base + 3 controlled creatures)", got)
	}

	expireCleanupDurations(g)

	if got := effectivePower(g, source); got != 2 {
		t.Fatalf("effective power after cleanup = %d, want 2 (buff expired)", got)
	}
}

// TestEventPermanentDynamicPumpScalesWithBoard proves that a triggered pump of
// the triggering permanent ("it gets +X/+X until end of turn, where X is the
// number of creatures you control.") resolves the event permanent reference and
// applies the computed +X/+X to it.
func TestEventPermanentDynamicPumpScalesWithBoard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	enchantment := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Pump Enchantment",
		Types: []types.Card{types.Enchantment},
	}})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	addCombatCreaturePermanentWithPower(g, game.Player1, 2)

	obj := &game.StackObject{
		Kind:            game.StackTriggeredAbility,
		SourceID:        enchantment.ObjectID,
		SourceCardID:    enchantment.CardInstanceID,
		Controller:      game.Player1,
		HasTriggerEvent: true,
		TriggerEvent:    game.Event{PermanentID: attacker.ObjectID},
	}
	count := countControlledCreatures()
	resolveInstruction(engine, g, obj, game.ModifyPT{
		Object:         game.EventPermanentReference(),
		PowerDelta:     game.Dynamic(count),
		ToughnessDelta: game.Dynamic(count),
		Duration:       game.DurationUntilEndOfTurn,
	}, &TurnLog{})

	if got := effectivePower(g, attacker); got != 4 {
		t.Fatalf("effective power = %d, want 4 (2 base + 2 controlled creatures)", got)
	}
	if got, _ := effectiveToughness(g, attacker); got != 4 {
		t.Fatalf("effective toughness = %d, want 4 (2 base + 2 controlled creatures)", got)
	}
}
