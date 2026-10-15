package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LeonardoWorldlyWarrior is the card definition for Leonardo, Worldly Warrior.
//
// Type: Legendary Creature — Mutant Ninja Turtle
// Cost: {7}{W}
//
// Oracle text:
//
//	Affinity for creatures (This spell costs {1} less to cast for each creature you control.)
//	Double strike (This creature deals both first-strike and regular combat damage.)
var LeonardoWorldlyWarrior = newLeonardoWorldlyWarrior

func newLeonardoWorldlyWarrior() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Leonardo, Worldly Warrior",
			ManaCost: opt.Val(cost.Mana{
				cost.O(7),
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Mutant, types.Ninja, types.Turtle},
			Power:      opt.Val(game.PT{Value: 5}),
			Toughness:  opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This spell costs {1} less to cast for each creature you control.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
							CostModifier: game.CostModifier{
								Kind:               game.CostModifierSpell,
								PerObjectReduction: 1,
								CountSelection:     &game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou},
							},
						},
					},
				},
				game.DoubleStrikeStaticBody,
			},
			OracleText: `
			Affinity for creatures (This spell costs {1} less to cast for each creature you control.)
			Double strike (This creature deals both first-strike and regular combat damage.)
		`,
		},
	}
}
