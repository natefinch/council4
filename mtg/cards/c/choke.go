package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Choke is the card definition for Choke.
//
// Type: Enchantment
// Cost: {2}{G}
//
// Oracle text:
//
//	Islands don't untap during their controllers' untap steps.
var Choke = newChoke()

func newChoke() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Choke",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:              game.RuleEffectDoesntUntap,
							AffectedSelection: game.Selection{SubtypesAny: []types.Sub{types.Sub("Island")}},
						},
					},
				},
			},
			OracleText: `
			Islands don't untap during their controllers' untap steps.
		`,
		},
	}
}
