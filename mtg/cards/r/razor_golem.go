package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RazorGolem is the card definition for Razor Golem.
//
// Type: Artifact Creature — Golem
// Cost: {6}
//
// Oracle text:
//
//	Affinity for Plains (This spell costs {1} less to cast for each Plains you control.)
//	Vigilance
var RazorGolem = newRazorGolem

func newRazorGolem() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Razor Golem",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Golem},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This spell costs {1} less to cast for each Plains you control.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
							CostModifier: game.CostModifier{
								Kind:               game.CostModifierSpell,
								PerObjectReduction: 1,
								CountSelection:     &game.Selection{SubtypesAny: []types.Sub{types.Plains}, Controller: game.ControllerYou},
							},
						},
					},
				},
				game.VigilanceStaticBody,
			},
			OracleText: `
			Affinity for Plains (This spell costs {1} less to cast for each Plains you control.)
			Vigilance
		`,
		},
	}
}
