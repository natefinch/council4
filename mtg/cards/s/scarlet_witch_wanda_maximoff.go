package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ScarletWitchWandaMaximoff is the card definition for Scarlet Witch, Wanda Maximoff.
//
// Type: Legendary Creature — Mutant Warlock Hero
// Cost: {2}{R}
//
// Oracle text:
//
//	Menace (This creature can't be blocked except by two or more creatures.)
var ScarletWitchWandaMaximoff = newScarletWitchWandaMaximoff

func newScarletWitchWandaMaximoff() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Scarlet Witch, Wanda Maximoff",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors:     []color.Color{color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Mutant, types.Warlock, types.Hero},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.MenaceStaticBody,
			},
			OracleText: `
			Menace (This creature can't be blocked except by two or more creatures.)
		`,
		},
	}
}
