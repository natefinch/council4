package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GeistFueledScarecrow is the card definition for Geist-Fueled Scarecrow.
//
// Type: Artifact Creature — Scarecrow
// Cost: {4}
//
// Oracle text:
//
//	Creature spells you cast cost {1} more to cast.
var GeistFueledScarecrow = newGeistFueledScarecrow

func newGeistFueledScarecrow() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Geist-Fueled Scarecrow",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Scarecrow},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Creature spells you cast cost {1} more to cast.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							Duration:       game.DurationPermanent,
							AffectedPlayer: game.PlayerYou,
							CostModifier: game.CostModifier{
								Kind:            game.CostModifierSpell,
								GenericIncrease: 1,
								CardSelection:   game.Selection{RequiredTypes: []types.Card{types.Creature}},
							},
						},
					},
				},
			},
			OracleText: `
			Creature spells you cast cost {1} more to cast.
		`,
		},
	}
}
