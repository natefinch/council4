package z

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ZurgoSVanguard is the card definition for Zurgo's Vanguard.
//
// Type: Creature — Dog Soldier
// Cost: {2}{R}
//
// Oracle text:
//
//	Mobilize 1 (Whenever this creature attacks, create a tapped and attacking 1/1 red Warrior creature token. Sacrifice it at the beginning of the next end step.)
//	Zurgo's Vanguard's power is equal to the number of creatures you control.
var ZurgoSVanguard = newZurgoSVanguard

func newZurgoSVanguard() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Zurgo's Vanguard",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors:       []color.Color{color.Red},
			Types:        []types.Card{types.Creature},
			Subtypes:     []types.Sub{types.Dog, types.Soldier},
			Power:        opt.Val(game.PT{IsStar: true}),
			Toughness:    opt.Val(game.PT{Value: 3}),
			DynamicPower: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerCreatureCount}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.MobilizeTriggeredBody(game.MobilizeAmount{Fixed: 1}),
			},
			OracleText: `
			Mobilize 1 (Whenever this creature attacks, create a tapped and attacking 1/1 red Warrior creature token. Sacrifice it at the beginning of the next end step.)
			Zurgo's Vanguard's power is equal to the number of creatures you control.
		`,
		},
	}
}
