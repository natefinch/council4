package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// TestSacrificedCostReferenceMillReadsLastKnownPower proves that a mill amount
// keyed to the sacrificed-cost permanent ("Sacrifice a creature: Target player
// mills cards equal to the sacrificed creature's power.") reads the sacrificed
// creature's last-known power after it has left the battlefield to pay the
// cost.
func TestSacrificedCostReferenceMillReadsLastKnownPower(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	sacrificed := addCombatCreaturePermanentWithPower(g, game.Player1, 4)
	for range 6 {
		addLibraryCard(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
			Name:  "Victim Card",
			Types: []types.Card{types.Land},
		}})
	}
	// Record last-known information, then remove the creature from the
	// battlefield to model it having been sacrificed to pay the ability cost.
	snapshot := snapshotPermanent(g, sacrificed, zone.Battlefield)
	rememberLastKnown(g, &snapshot)
	removePermanentFromBattlefield(g, sacrificed.ObjectID)

	obj := &game.StackObject{
		Kind:                game.StackActivatedAbility,
		Controller:          game.Player1,
		Targets:             []game.Target{game.PlayerTarget(game.Player2)},
		SacrificedAsCostIDs: []id.ID{sacrificed.ObjectID},
	}
	resolveInstruction(engine, g, obj, game.Mill{
		Amount: game.Dynamic(game.DynamicAmount{
			Kind:       game.DynamicAmountObjectPower,
			Multiplier: 1,
			Object:     game.SacrificedCostReference(),
		}),
		Player: game.TargetPlayerReference(0),
	}, &TurnLog{})

	if got := g.Players[game.Player2].Library.Size(); got != 2 {
		t.Fatalf("Player2 library size = %d, want 2 (6 - 4 milled)", got)
	}
	if got := g.Players[game.Player2].Graveyard.Size(); got != 4 {
		t.Fatalf("Player2 graveyard size = %d, want 4 milled", got)
	}
}
