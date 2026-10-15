package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GryffRider is the card definition for Gryff Rider.
//
// Type: Creature — Human Knight
// Cost: {2}{W}
//
// Oracle text:
//
//	Flying
//	Training (Whenever this creature attacks with another creature with greater power, put a +1/+1 counter on this creature.)
var GryffRider = newGryffRider

func newGryffRider() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Gryff Rider",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Knight},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TrainingTriggeredBody,
			},
			OracleText: `
			Flying
			Training (Whenever this creature attacks with another creature with greater power, put a +1/+1 counter on this creature.)
		`,
		},
	}
}
