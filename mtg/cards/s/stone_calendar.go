package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// StoneCalendar is the card definition for Stone Calendar.
//
// Type: Artifact
// Cost: {5}
//
// Oracle text:
//
//	Spells you cast cost {1} less to cast.
var StoneCalendar = newStoneCalendar

func newStoneCalendar() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Stone Calendar",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
			}),
			Types: []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Spells you cast cost {1} less to cast.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							Duration:       game.DurationPermanent,
							AffectedPlayer: game.PlayerYou,
							CostModifier: game.CostModifier{
								Kind:             game.CostModifierSpell,
								GenericReduction: 1,
							},
						},
					},
				},
			},
			OracleText: `
			Spells you cast cost {1} less to cast.
		`,
		},
	}
}
