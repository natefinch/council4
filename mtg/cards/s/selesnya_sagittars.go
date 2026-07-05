package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SelesnyaSagittars is the card definition for Selesnya Sagittars.
//
// Type: Creature — Elf Archer
// Cost: {3}{G}{W}
//
// Oracle text:
//
//	Reach (This creature can block creatures with flying.)
//	This creature can block an additional creature each combat.
var SelesnyaSagittars = newSelesnyaSagittars()

func newSelesnyaSagittars() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name: "Selesnya Sagittars",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
				cost.W,
			}),
			Colors:    []color.Color{color.Green, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elf, types.Archer},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.ReachStaticBody,
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:                 game.RuleEffectCanBlockAdditional,
							AffectedSource:       true,
							AdditionalBlockCount: 1,
						},
					},
				},
			},
			OracleText: `
			Reach (This creature can block creatures with flying.)
			This creature can block an additional creature each combat.
		`,
		},
	}
}
