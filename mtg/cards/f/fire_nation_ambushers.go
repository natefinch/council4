package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FireNationAmbushers is the card definition for Fire Nation Ambushers.
//
// Type: Creature — Human Soldier
// Cost: {2}{B}
//
// Oracle text:
//
//	Flash (You may cast this spell any time you could cast an instant.)
var FireNationAmbushers = newFireNationAmbushers

func newFireNationAmbushers() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Fire Nation Ambushers",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Soldier},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
			},
			OracleText: `
			Flash (You may cast this spell any time you could cast an instant.)
		`,
		},
	}
}
