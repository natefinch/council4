package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CallTheCoppercoats is the card definition for Call the Coppercoats.
//
// Type: Instant
// Cost: {2}{W}
//
// Oracle text:
//
//	Strive — This spell costs {1}{W} more to cast for each target beyond the first.
//	Choose any number of target opponents. Create X 1/1 white Human Soldier creature tokens, where X is the number of creatures those opponents control.
var CallTheCoppercoats = newCallTheCoppercoats

func newCallTheCoppercoats() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Call the Coppercoats",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Strive — This spell costs {1}{W} more to cast for each target beyond the first.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
							CostModifier: game.CostModifier{
								Kind: game.CostModifierSpell,
								PerTargetBeyondFirstIncrease: cost.Mana{
									cost.O(1),
									cost.W,
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
						Constraint: "any number of target opponents",
						Allow:      game.TargetAllowPlayer,
						Selection:  opt.Val(game.Selection{Player: game.PlayerOpponent}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountCountSelector,
								Multiplier: 1,
								Group:      game.PlayerGroupControlledGroup(game.TargetedPlayersReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							}),
							Source: game.TokenDef(callTheCoppercoatsToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Strive — This spell costs {1}{W} more to cast for each target beyond the first.
			Choose any number of target opponents. Create X 1/1 white Human Soldier creature tokens, where X is the number of creatures those opponents control.
		`,
		},
	}
}

var callTheCoppercoatsToken = newCallTheCoppercoatsToken()

func newCallTheCoppercoatsToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Human Soldier",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Soldier},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
