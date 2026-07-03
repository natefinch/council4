package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TitanOfIndustry is the card definition for Titan of Industry.
//
// Type: Creature — Elemental
// Cost: {4}{G}{G}{G}
//
// Oracle text:
//
//	Reach, trample
//	When this creature enters, choose two —
//	• Destroy target artifact or enchantment.
//	• Target player gains 5 life.
//	• Create a 4/4 green Rhino Warrior creature token.
//	• Put a shield counter on a creature you control.
var TitanOfIndustry = newTitanOfIndustry()

func newTitanOfIndustry() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Titan of Industry",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elemental},
			Power:     opt.Val(game.PT{Value: 7}),
			Toughness: opt.Val(game.PT{Value: 7}),
			StaticAbilities: []game.StaticAbility{
				game.ReachStaticBody,
				game.TrampleStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Text: "Destroy target artifact or enchantment.",
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target artifact or enchantment",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Enchantment}}),
									},
								},
								Sequence: []game.Instruction{
									{
										Primitive: game.Destroy{
											Object: game.TargetPermanentReference(0),
										},
									},
								},
							},
							game.Mode{
								Text: "Target player gains 5 life.",
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "Target player",
										Allow:      game.TargetAllowPlayer,
									},
								},
								Sequence: []game.Instruction{
									{
										Primitive: game.GainLife{
											Amount: game.Fixed(5),
											Player: game.TargetPlayerReference(0),
										},
									},
								},
							},
							game.Mode{
								Text: "Create a 4/4 green Rhino Warrior creature token.",
								Sequence: []game.Instruction{
									{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(titanOfIndustryToken),
										},
									},
								},
							},
							game.Mode{
								Text: "Put a shield counter on a creature you control.",
								Sequence: []game.Instruction{
									{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
											ChooseOne:   true,
											CounterKind: counter.Shield,
										},
									},
								},
							},
						},
						MinModes: 2,
						MaxModes: 2,
					},
				},
			},
			OracleText: `
			Reach, trample
			When this creature enters, choose two —
			• Destroy target artifact or enchantment.
			• Target player gains 5 life.
			• Create a 4/4 green Rhino Warrior creature token.
			• Put a shield counter on a creature you control.
		`,
		},
	}
}

var titanOfIndustryToken = newTitanOfIndustryToken()

func newTitanOfIndustryToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Rhino Warrior",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Rhino, types.Warrior},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
		},
	}
}
