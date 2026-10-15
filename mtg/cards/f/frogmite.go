package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Frogmite is the card definition for Frogmite.
//
// Type: Artifact Creature — Frog
// Cost: {4}
//
// Oracle text:
//
//	Affinity for artifacts (This spell costs {1} less to cast for each artifact you control.)
var Frogmite = newFrogmite

func newFrogmite() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Frogmite",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Frog},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This spell costs {1} less to cast for each artifact you control.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
							CostModifier: game.CostModifier{
								Kind:               game.CostModifierSpell,
								PerObjectReduction: 1,
								CountSelection:     &game.Selection{RequiredTypes: []types.Card{types.Artifact}, Controller: game.ControllerYou},
							},
						},
					},
				},
			},
			OracleText: `
			Affinity for artifacts (This spell costs {1} less to cast for each artifact you control.)
		`,
		},
	}
}
