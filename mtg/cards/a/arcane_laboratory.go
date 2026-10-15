package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ArcaneLaboratory is the card definition for Arcane Laboratory.
//
// Type: Enchantment
// Cost: {2}{U}
//
// Oracle text:
//
//	Each player can't cast more than one spell each turn.
var ArcaneLaboratory = newArcaneLaboratory

func newArcaneLaboratory() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Arcane Laboratory",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Each player can't cast more than one spell each turn.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:             game.RuleEffectCastLimitPerTurn,
							Duration:         game.DurationPermanent,
							CastLimitPerTurn: 1,
						},
					},
				},
			},
			OracleText: `
			Each player can't cast more than one spell each turn.
		`,
		},
	}
}
