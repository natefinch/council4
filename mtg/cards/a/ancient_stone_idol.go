package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AncientStoneIdol is the card definition for Ancient Stone Idol.
//
// Type: Artifact Creature — Golem
// Cost: {10}
//
// Oracle text:
//
//	Flash
//	This spell costs {1} less to cast for each attacking creature.
//	Trample
//	When this creature dies, create a 6/12 colorless Construct artifact creature token with trample.
var AncientStoneIdol = newAncientStoneIdol()

func newAncientStoneIdol() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Ancient Stone Idol",
			ManaCost: opt.Val(cost.Mana{
				cost.O(10),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Golem},
			Power:     opt.Val(game.PT{Value: 12}),
			Toughness: opt.Val(game.PT{Value: 12}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedSource: true,
							CostModifier: game.CostModifier{
								Kind:               game.CostModifierSpell,
								PerObjectReduction: 1,
								CountSelection:     &game.Selection{RequiredTypes: []types.Card{types.Creature}, CombatState: game.CombatStateAttacking},
							},
						},
					},
				},
				game.TrampleStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentDied,
							Source:           game.TriggerSourceSelf,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(ancientStoneIdolToken),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flash
			This spell costs {1} less to cast for each attacking creature.
			Trample
			When this creature dies, create a 6/12 colorless Construct artifact creature token with trample.
		`,
		},
	}
}

var ancientStoneIdolToken = newAncientStoneIdolToken()

func newAncientStoneIdolToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Construct",
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Construct},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 12}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
			},
		},
	}
}
