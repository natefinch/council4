package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GoblinChampion is the card definition for Goblin Champion.
//
// Type: Creature — Goblin Warrior
// Cost: {R}
//
// Oracle text:
//
//	Haste
//	Exalted (Whenever a creature you control attacks alone, that creature gets +1/+1 until end of turn.)
var GoblinChampion = newGoblinChampion

func newGoblinChampion() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Goblin Champion",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Goblin, types.Warrior},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.HasteStaticBody,
				game.ExaltedStaticBody,
			},
			OracleText: `
			Haste
			Exalted (Whenever a creature you control attacks alone, that creature gets +1/+1 until end of turn.)
		`,
		},
	}
}
