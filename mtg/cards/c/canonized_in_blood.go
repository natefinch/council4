package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// CanonizedInBlood is the card definition for Canonized in Blood.
//
// Type: Enchantment
// Cost: {1}{B}
//
// Oracle text:
//
//	At the beginning of your end step, if you descended this turn, put a +1/+1 counter on target creature you control. (You descended if a permanent card was put into your graveyard from anywhere.)
//	{5}{B}{B}, Sacrifice this enchantment: Create a 4/3 white and black Vampire Demon creature token with flying.
var CanonizedInBlood = newCanonizedInBlood()

func newCanonizedInBlood() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Canonized in Blood",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Enchantment},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{5}{B}{B}, Sacrifice this enchantment: Create a 4/3 white and black Vampire Demon creature token with flying.",
					ManaCost: opt.Val(cost.Mana{cost.O(5), cost.B, cost.B}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:   cost.AdditionalSacrificeSource,
							Text:   "Sacrifice this enchantment",
							Amount: 1,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(canonizedInBloodToken),
								},
							},
						},
					}.Ability(),
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepEnd,
						},
						InterveningIf: "if you descended this turn",
						InterveningCondition: opt.Val(game.Condition{
							EventHistory: opt.Val(game.EventHistoryCondition{Pattern: game.TriggerPattern{
								Event:            game.EventZoneChanged,
								Player:           game.TriggerPlayerYou,
								MatchToZone:      true,
								ToZone:           zone.Graveyard,
								SubjectSelection: game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Battle, types.Creature, types.Enchantment, types.Land, types.Planeswalker}, NonToken: true},
							}, Window: game.EventHistoryCurrentTurn}),
						}),
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature you control",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.TargetPermanentReference(0),
									CounterKind: counter.PlusOnePlusOne,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			At the beginning of your end step, if you descended this turn, put a +1/+1 counter on target creature you control. (You descended if a permanent card was put into your graveyard from anywhere.)
			{5}{B}{B}, Sacrifice this enchantment: Create a 4/3 white and black Vampire Demon creature token with flying.
		`,
		},
	}
}

var canonizedInBloodToken = newCanonizedInBloodToken()

func newCanonizedInBloodToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Vampire Demon",
			Colors:    []color.Color{color.White, color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Vampire, types.Demon},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
		},
	}
}
