package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HourOfNeed is the card definition for Hour of Need.
//
// Type: Instant
// Cost: {2}{U}
//
// Oracle text:
//
//	Strive — This spell costs {1}{U} more to cast for each target beyond the first.
//	Exile any number of target creatures. For each creature exiled this way, its controller creates a 4/4 blue Sphinx creature token with flying.
var HourOfNeed = newHourOfNeed

func newHourOfNeed() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Hour of Need",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Strive — This spell costs {1}{U} more to cast for each target beyond the first.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
							CostModifier: game.CostModifier{
								Kind: game.CostModifierSpell,
								PerTargetBeyondFirstIncrease: cost.Mana{
									cost.O(1),
									cost.U,
								},
							},
						},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 0,
						MaxTargets: 99,
						Constraint: "any number of target creatures",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.RemoveTargetsForToken{
							Exile:     true,
							LinkedKey: game.LinkedKey("removed-targets-for-token"),
						},
					},
					{
						Primitive: game.CreateTokenForEachDestroyed{
							Source:    game.TokenDef(hourOfNeedToken),
							LinkedKey: game.LinkedKey("removed-targets-for-token"),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Strive — This spell costs {1}{U} more to cast for each target beyond the first.
			Exile any number of target creatures. For each creature exiled this way, its controller creates a 4/4 blue Sphinx creature token with flying.
		`,
		},
	}
}

var hourOfNeedToken = newHourOfNeedToken()

func newHourOfNeedToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Sphinx",
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Sphinx},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
		},
	}
}
