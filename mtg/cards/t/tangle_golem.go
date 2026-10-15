package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TangleGolem is the card definition for Tangle Golem.
//
// Type: Artifact Creature — Golem
// Cost: {7}
//
// Oracle text:
//
//	Affinity for Forests (This spell costs {1} less to cast for each Forest you control.)
var TangleGolem = newTangleGolem

func newTangleGolem() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Tangle Golem",
			ManaCost: opt.Val(cost.Mana{
				cost.O(7),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Golem},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This spell costs {1} less to cast for each Forest you control.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
							CostModifier: game.CostModifier{
								Kind:               game.CostModifierSpell,
								PerObjectReduction: 1,
								CountSelection:     &game.Selection{SubtypesAny: []types.Sub{types.Forest}, Controller: game.ControllerYou},
							},
						},
					},
				},
			},
			OracleText: `
			Affinity for Forests (This spell costs {1} less to cast for each Forest you control.)
		`,
		},
	}
}
