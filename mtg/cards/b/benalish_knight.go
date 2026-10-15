package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BenalishKnight is the card definition for Benalish Knight.
//
// Type: Creature — Human Knight
// Cost: {2}{W}
//
// Oracle text:
//
//	Flash (You may cast this spell any time you could cast an instant.)
//	First strike (This creature deals combat damage before creatures without first strike.)
var BenalishKnight = newBenalishKnight

func newBenalishKnight() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Benalish Knight",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Knight},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
				game.FirstStrikeStaticBody,
			},
			OracleText: `
			Flash (You may cast this spell any time you could cast an instant.)
			First strike (This creature deals combat damage before creatures without first strike.)
		`,
		},
	}
}
