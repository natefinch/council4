package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestDynamicAmountOpponentControllingCount proves that the
// DynamicAmountOpponentControllingCount kind (Summon: Yojimbo chapter IV's
// "the number of opponents who control a creature with power 4 or greater")
// counts one per opponent who controls at least one permanent matching the
// per-opponent Group, and ignores both the controller's own qualifying
// permanents and eliminated opponents.
func TestDynamicAmountOpponentControllingCount(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	bigCreature := func(name string, power int) *game.CardDef {
		return &game.CardDef{CardFace: game.CardFace{
			Name:  name,
			Types: []types.Card{types.Creature},
			Power: opt.Val(game.PT{Value: power}),
		}}
	}
	// The controller's own 9-power creature must not contribute.
	addCombatPermanent(g, game.Player1, bigCreature("Mine", 9))
	// Player2 controls a power-4 creature: qualifies.
	addCombatPermanent(g, game.Player2, bigCreature("Theirs Big", 4))
	// Player3 controls only a power-2 creature: does not qualify.
	addCombatPermanent(g, game.Player3, bigCreature("Theirs Small", 2))
	// Player4 controls a power-7 creature but is eliminated: does not count.
	addCombatPermanent(g, game.Player4, bigCreature("Dead Big", 7))
	g.Players[game.Player4].Eliminated = true

	obj := &game.StackObject{Controller: game.Player1}
	amount := game.DynamicAmount{
		Kind: game.DynamicAmountOpponentControllingCount,
		Group: game.BattlefieldGroup(game.Selection{
			RequiredTypes: []types.Card{types.Creature},
			Controller:    game.ControllerYou,
			Power:         opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 4}),
		}),
	}
	if got := dynamicAmountValue(g, obj, game.Player1, amount); got != 1 {
		t.Fatalf("opponents controlling a power>=4 creature = %d, want 1", got)
	}

	// Give Player3 a power-5 creature; now two opponents qualify.
	addCombatPermanent(g, game.Player3, bigCreature("Theirs Now Big", 5))
	if got := dynamicAmountValue(g, obj, game.Player1, amount); got != 2 {
		t.Fatalf("opponents controlling a power>=4 creature = %d, want 2", got)
	}
}
