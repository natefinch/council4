package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// VodalianHexcatcher is the card definition for Vodalian Hexcatcher.
//
// Type: Creature — Merfolk Wizard
// Cost: {1}{U}
//
// Oracle text:
//
//	Flash
//	Other Merfolk you control get +1/+1.
//	Sacrifice a Merfolk: Counter target noncreature spell unless its controller pays {1}.
var VodalianHexcatcher = newVodalianHexcatcher

func newVodalianHexcatcher() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Vodalian Hexcatcher",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Merfolk, types.Wizard},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							Group:          game.ObjectControlledGroupExcluding(game.SourcePermanentReference(), game.Selection{SubtypesAny: []types.Sub{types.Sub("Merfolk")}}, game.SourcePermanentReference()),
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Sacrifice a Merfolk: Counter target noncreature spell unless its controller pays {1}.",
					AdditionalCosts: []cost.Additional{
						{
							Kind:        cost.AdditionalSacrifice,
							Text:        "Sacrifice a Merfolk",
							Amount:      1,
							SubtypesAny: cost.SubtypeSet{types.Merfolk},
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target noncreature spell",
								Allow:      game.TargetAllowStackObject,
								Predicate: game.TargetPredicate{
									ExcludedSpellCardTypes: []types.Card{types.Creature},
									StackObjectKinds:       []game.StackObjectKind{game.StackSpell},
								},
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Pay{
									Payment: game.ResolutionPayment{
										Prompt: "Pay {1}?",
										Payer:  opt.Val(game.ObjectControllerReference(game.TargetStackObjectReference(0))),
										ManaCost: opt.Val(cost.Mana{
											cost.O(1),
										}),
									},
								},
								PublishResult: game.ResultKey("unless-paid"),
							},
							{
								Primitive: game.CounterObject{
									Object: game.TargetStackObjectReference(0),
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "unless-paid",
									Succeeded: game.TriFalse,
								}),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flash
			Other Merfolk you control get +1/+1.
			Sacrifice a Merfolk: Counter target noncreature spell unless its controller pays {1}.
		`,
		},
	}
}
