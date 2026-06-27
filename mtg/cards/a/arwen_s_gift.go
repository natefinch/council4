package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ArwenSGift is the card definition for Arwen's Gift.
//
// Type: Sorcery
// Cost: {3}{U}
//
// Oracle text:
//
//	This spell costs {1} less to cast if you control two or more legendary creatures.
//	Scry 2, then draw two cards.
var ArwenSGift = newArwenSGift()

func newArwenSGift() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Arwen's Gift",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedSource: true,
							CostModifier: game.CostModifier{
								Kind:             game.CostModifierSpell,
								GenericReduction: 1,
								ReductionCondition: opt.Val(game.Condition{
									ControlsMatching: opt.Val(game.SelectionCount{
										Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, Supertypes: []types.Super{types.Legendary}},
										MinCount:  2,
									}),
								}),
							},
						},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Scry{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
					},
					{
						Primitive: game.Draw{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			This spell costs {1} less to cast if you control two or more legendary creatures.
			Scry 2, then draw two cards.
		`,
		},
	}
}
