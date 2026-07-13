package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HeraldOfSerra is the card definition for Herald of Serra.
//
// Type: Creature — Angel
// Cost: {2}{W}{W}
//
// Oracle text:
//
//	Flying, vigilance
//	Echo {2}{W}{W} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
var HeraldOfSerra = newHeraldOfSerra

func newHeraldOfSerra() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Herald of Serra",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Angel},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.VigilanceStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.EchoTriggeredAbility(cost.Mana{cost.O(2), cost.W, cost.W}),
			},
			OracleText: `
			Flying, vigilance
			Echo {2}{W}{W} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
		`,
		},
	}
}
