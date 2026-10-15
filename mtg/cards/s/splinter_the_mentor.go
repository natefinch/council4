package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SplinterTheMentor is the card definition for Splinter, the Mentor.
//
// Type: Legendary Creature — Mutant Ninja Rat
// Cost: {1}{B}
//
// Oracle text:
//
//	Menace
//	Whenever Splinter or another nontoken creature you control leaves the battlefield, create a Mutagen token.
//	Partner—Character select (You can have two commanders if both have this ability.)
var SplinterTheMentor = newSplinterTheMentor

func newSplinterTheMentor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Splinter, the Mentor",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors:     []color.Color{color.Black},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Mutant, types.Ninja, types.Rat},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.MenaceStaticBody,
				game.PartnerStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever Splinter or another nontoken creature you control leaves the battlefield, create a Mutagen token.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                  game.EventZoneChanged,
							Controller:             game.TriggerControllerYou,
							SubjectSelection:       game.Selection{RequiredTypes: []types.Card{types.Creature}, NonToken: true},
							SubjectSelectionOrSelf: true,
							MatchFromZone:          true,
							FromZone:               zone.Battlefield,
							StackObjectKind:        game.StackSpell,
							CounterKind:            counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(splinterTheMentorToken),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			OracleText: `
			Menace
			Whenever Splinter or another nontoken creature you control leaves the battlefield, create a Mutagen token.
			Partner—Character select (You can have two commanders if both have this ability.)
		`,
		},
	}
}

var splinterTheMentorToken = newSplinterTheMentorToken()

func newSplinterTheMentorToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Mutagen",
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Mutagen},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{1}, {T}, Sacrifice this token: Put a +1/+1 counter on target creature. Activate only as a sorcery.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(1),
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
						cost.Additional{
							Kind:        cost.AdditionalSacrificeSource,
							Text:        "Sacrifice this token",
							Amount:      1,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Timing:         game.SorceryOnly,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target creature",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Object:      game.TargetPermanentReference(0),
											CounterKind: counter.PlusOnePlusOne,
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
		},
	}
}
