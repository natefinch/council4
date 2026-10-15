package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AngelicObserver is the card definition for Angelic Observer.
//
// Type: Creature — Angel Advisor
// Cost: {5}{W}
//
// Oracle text:
//
//	Affinity for Citizens (This spell costs {1} less to cast for each Citizen you control.)
//	Flying
var AngelicObserver = newAngelicObserver

func newAngelicObserver() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Angelic Observer",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Angel, types.Advisor},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This spell costs {1} less to cast for each Citizen you control.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
							CostModifier: game.CostModifier{
								Kind:               game.CostModifierSpell,
								PerObjectReduction: 1,
								CountSelection:     &game.Selection{SubtypesAny: []types.Sub{types.Citizen}, Controller: game.ControllerYou},
							},
						},
					},
				},
				game.FlyingStaticBody,
			},
			OracleText: `
			Affinity for Citizens (This spell costs {1} less to cast for each Citizen you control.)
			Flying
		`,
		},
	}
}
