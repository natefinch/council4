package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// EmberethVeteran is the card definition for Embereth Veteran.
//
// Type: Creature — Human Knight
// Cost: {R}
//
// Oracle text:
//
//	{1}, Sacrifice this creature: Create a Young Hero Role token attached to another target creature. (If you control another Role on it, put that one into the graveyard. Enchanted creature has "Whenever this creature attacks, if its toughness is 3 or less, put a +1/+1 counter on it.")
var EmberethVeteran = newEmberethVeteran

func newEmberethVeteran() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Embereth Veteran",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Knight},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{1}, Sacrifice this creature: Create a Young Hero Role token attached to another target creature. (If you control another Role on it, put that one into the graveyard. Enchanted creature has \"Whenever this creature attacks, if its toughness is 3 or less, put a +1/+1 counter on it.\")",
					ManaCost: opt.Val(cost.Mana{cost.O(1)}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:   cost.AdditionalSacrificeSource,
							Text:   "Sacrifice this creature",
							Amount: 1,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "another target creature",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, ExcludeSource: true}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount:          game.Fixed(1),
									Source:          game.TokenDef(emberethVeteranToken),
									EntryAttachedTo: opt.Val(game.TargetObjectReference(0)),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{1}, Sacrifice this creature: Create a Young Hero Role token attached to another target creature. (If you control another Role on it, put that one into the graveyard. Enchanted creature has "Whenever this creature attacks, if its toughness is 3 or less, put a +1/+1 counter on it.")
		`,
		},
	}
}

var emberethVeteranToken = newEmberethVeteranToken()

func newEmberethVeteranToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Young Hero Role",
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura, types.Role},
			StaticAbilities: []game.StaticAbility{
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
				}),
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
							AddAbilities: []game.Ability{
								new(game.TriggeredAbility{
									Trigger: game.TriggerCondition{
										Type: game.TriggerWhenever,
										Pattern: game.TriggerPattern{
											Event:  game.EventAttackerDeclared,
											Source: game.TriggerSourceSelf,
										},
										InterveningIf: "if its toughness is 3 or less",
										InterveningCondition: opt.Val(game.Condition{
											Object:        opt.Val(game.EventPermanentReference()),
											ObjectMatches: opt.Val(game.Selection{Toughness: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 3})}),
										}),
									},
									Content: game.Mode{
										Sequence: []game.Instruction{
											{
												Primitive: game.AddCounter{
													Amount:      game.Fixed(1),
													Object:      game.EventPermanentReference(),
													CounterKind: counter.PlusOnePlusOne,
												},
											},
										},
									}.Ability(),
								}),
							},
						},
					},
				},
			},
			OracleText: `
			Enchant creature
			Enchanted creature has "Whenever this creature attacks, if its toughness is 3 or less, put a +1/+1 counter on it."
		`,
		},
	}
}
