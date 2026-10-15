package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Human is the card definition for Human.
//
// Type: Token Creature — Human
//
// Oracle text:
//   Doctor spells you cast cost {1} less to cast.

// HumanToken8e0677013b0b4985a8f9fd634b6b8e78 is the card definition for Human.
var HumanToken8e0677013b0b4985a8f9fd634b6b8e78 = newHumanToken8e0677013b0b4985a8f9fd634b6b8e78()

func newHumanToken8e0677013b0b4985a8f9fd634b6b8e78() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name:      "Human",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Doctor spells you cast cost {1} less to cast.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							Duration:       game.DurationPermanent,
							AffectedPlayer: game.PlayerYou,
							CostModifier: game.CostModifier{
								Kind:             game.CostModifierSpell,
								GenericReduction: 1,
								CardSelection:    game.Selection{SubtypesAny: []types.Sub{types.Doctor}},
							},
						},
					},
				},
			},
			OracleText: `
			Doctor spells you cast cost {1} less to cast.
		`,
		},
	}
}
