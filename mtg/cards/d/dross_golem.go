package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DrossGolem is the card definition for Dross Golem.
//
// Type: Artifact Creature — Golem
// Cost: {5}
//
// Oracle text:
//
//	Affinity for Swamps (This spell costs {1} less to cast for each Swamp you control.)
//	Fear (This creature can't be blocked except by artifact creatures and/or black creatures.)
var DrossGolem = newDrossGolem

func newDrossGolem() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Dross Golem",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Golem},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This spell costs {1} less to cast for each Swamp you control.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
							CostModifier: game.CostModifier{
								Kind:               game.CostModifierSpell,
								PerObjectReduction: 1,
								CountSelection:     &game.Selection{SubtypesAny: []types.Sub{types.Swamp}, Controller: game.ControllerYou},
							},
						},
					},
				},
				game.FearStaticBody,
			},
			OracleText: `
			Affinity for Swamps (This spell costs {1} less to cast for each Swamp you control.)
			Fear (This creature can't be blocked except by artifact creatures and/or black creatures.)
		`,
		},
	}
}
