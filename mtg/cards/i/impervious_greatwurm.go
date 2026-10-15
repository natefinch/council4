package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ImperviousGreatwurm is the card definition for Impervious Greatwurm.
//
// Type: Creature — Wurm
// Cost: {7}{G}{G}{G}
//
// Oracle text:
//
//	Convoke (Your creatures can help cast this spell. Each creature you tap while casting this spell pays for {1} or one mana of that creature's color.)
//	Indestructible
var ImperviousGreatwurm = newImperviousGreatwurm

func newImperviousGreatwurm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Impervious Greatwurm",
			ManaCost: opt.Val(cost.Mana{
				cost.O(7),
				cost.G,
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Wurm},
			Power:     opt.Val(game.PT{Value: 16}),
			Toughness: opt.Val(game.PT{Value: 16}),
			StaticAbilities: []game.StaticAbility{
				game.ConvokeStaticBody,
				game.IndestructibleStaticBody,
			},
			OracleText: `
			Convoke (Your creatures can help cast this spell. Each creature you tap while casting this spell pays for {1} or one mana of that creature's color.)
			Indestructible
		`,
		},
	}
}
