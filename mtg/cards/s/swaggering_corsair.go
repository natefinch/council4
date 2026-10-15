package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SwaggeringCorsair is the card definition for Swaggering Corsair.
//
// Type: Creature — Human Pirate
// Cost: {2}{R}
//
// Oracle text:
//
//	Raid — This creature enters with a +1/+1 counter on it if you attacked this turn.
var SwaggeringCorsair = newSwaggeringCorsair

func newSwaggeringCorsair() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Swaggering Corsair",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Pirate},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersIfReplacement("Raid — This creature enters with a +1/+1 counter on it if you attacked this turn.", &game.Condition{
					EventHistory: opt.Val(game.EventHistoryCondition{Pattern: game.TriggerPattern{
						Event:           game.EventAttackerDeclared,
						Controller:      game.TriggerControllerYou,
						StackObjectKind: game.StackSpell,
						CounterKind:     counter.PlusOnePlusOne,
					}, Window: game.EventHistoryCurrentTurn}),
				}, game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 1}),
			},
			OracleText: `
			Raid — This creature enters with a +1/+1 counter on it if you attacked this turn.
		`,
		},
	}
}
