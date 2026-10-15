package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// OxiddaGolem is the card definition for Oxidda Golem.
//
// Type: Artifact Creature — Golem
// Cost: {6}
//
// Oracle text:
//
//	Affinity for Mountains (This spell costs {1} less to cast for each Mountain you control.)
//	Haste
var OxiddaGolem = newOxiddaGolem

func newOxiddaGolem() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Oxidda Golem",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Golem},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This spell costs {1} less to cast for each Mountain you control.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
							CostModifier: game.CostModifier{
								Kind:               game.CostModifierSpell,
								PerObjectReduction: 1,
								CountSelection:     &game.Selection{SubtypesAny: []types.Sub{types.Mountain}, Controller: game.ControllerYou},
							},
						},
					},
				},
				game.HasteStaticBody,
			},
			OracleText: `
			Affinity for Mountains (This spell costs {1} less to cast for each Mountain you control.)
			Haste
		`,
		},
	}
}
