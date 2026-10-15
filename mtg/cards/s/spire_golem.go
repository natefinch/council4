package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SpireGolem is the card definition for Spire Golem.
//
// Type: Artifact Creature — Golem
// Cost: {6}
//
// Oracle text:
//
//	Affinity for Islands (This spell costs {1} less to cast for each Island you control.)
//	Flying
var SpireGolem = newSpireGolem

func newSpireGolem() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Spire Golem",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Golem},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This spell costs {1} less to cast for each Island you control.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
							CostModifier: game.CostModifier{
								Kind:               game.CostModifierSpell,
								PerObjectReduction: 1,
								CountSelection:     &game.Selection{SubtypesAny: []types.Sub{types.Island}, Controller: game.ControllerYou},
							},
						},
					},
				},
				game.FlyingStaticBody,
			},
			OracleText: `
			Affinity for Islands (This spell costs {1} less to cast for each Island you control.)
			Flying
		`,
		},
	}
}
