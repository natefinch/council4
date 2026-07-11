package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SumalaRumblers is the card definition for Sumala Rumblers.
//
// Type: Creature — Wurm
// Cost: {2}{G/W}{G/W}
//
// Oracle text:
//
//	Sumala Rumblers's power is equal to the number of creatures you control.
//	Myriad (Whenever this creature attacks, for each opponent other than defending player, you may create a token copy that's tapped and attacking that player or a planeswalker they control. Exile the tokens at end of combat.)
var SumalaRumblers = newSumalaRumblers

func newSumalaRumblers() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name: "Sumala Rumblers",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.HybridMana(mana.G, mana.W),
				cost.HybridMana(mana.G, mana.W),
			}),
			Colors:       []color.Color{color.Green, color.White},
			Types:        []types.Card{types.Creature},
			Subtypes:     []types.Sub{types.Wurm},
			Power:        opt.Val(game.PT{IsStar: true}),
			Toughness:    opt.Val(game.PT{Value: 4}),
			DynamicPower: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerCreatureCount}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.MyriadTriggeredBody,
			},
			OracleText: `
			Sumala Rumblers's power is equal to the number of creatures you control.
			Myriad (Whenever this creature attacks, for each opponent other than defending player, you may create a token copy that's tapped and attacking that player or a planeswalker they control. Exile the tokens at end of combat.)
		`,
		},
	}
}
