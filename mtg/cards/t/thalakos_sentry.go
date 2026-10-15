package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ThalakosSentry is the card definition for Thalakos Sentry.
//
// Type: Creature — Thalakos Soldier
// Cost: {1}{U}
//
// Oracle text:
//
//	Shadow (This creature can block or be blocked by only creatures with shadow.)
var ThalakosSentry = newThalakosSentry

func newThalakosSentry() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Thalakos Sentry",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Thalakos, types.Soldier},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.ShadowStaticBody,
			},
			OracleText: `
			Shadow (This creature can block or be blocked by only creatures with shadow.)
		`,
		},
	}
}
