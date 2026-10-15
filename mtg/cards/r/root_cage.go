package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RootCage is the card definition for Root Cage.
//
// Type: Enchantment
// Cost: {1}{G}
//
// Oracle text:
//
//	Mercenaries don't untap during their controllers' untap steps.
var RootCage = newRootCage

func newRootCage() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Root Cage",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Mercenaries don't untap during their controllers' untap steps.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:              game.RuleEffectDoesntUntap,
							Duration:          game.DurationPermanent,
							AffectedSelection: game.Selection{SubtypesAny: []types.Sub{types.Mercenary}},
						},
					},
				},
			},
			OracleText: `
			Mercenaries don't untap during their controllers' untap steps.
		`,
		},
	}
}
