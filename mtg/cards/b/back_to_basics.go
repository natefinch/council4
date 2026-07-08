package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BackToBasics is the card definition for Back to Basics.
//
// Type: Enchantment
// Cost: {2}{U}
//
// Oracle text:
//
//	Nonbasic lands don't untap during their controllers' untap steps.
var BackToBasics = newBackToBasics

func newBackToBasics() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Back to Basics",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:              game.RuleEffectDoesntUntap,
							PermanentTypes:    []types.Card{types.Land},
							AffectedSelection: game.Selection{ExcludedSupertype: types.Basic},
						},
					},
				},
			},
			OracleText: `
			Nonbasic lands don't untap during their controllers' untap steps.
		`,
		},
	}
}
