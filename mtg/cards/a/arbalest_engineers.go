package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ArbalestEngineers is the card definition for Arbalest Engineers.
//
// Type: Creature — Human Artificer
// Cost: {1}{R}{G}
//
// Oracle text:
//
//	When this creature enters, choose one —
//	• This creature deals 1 damage to any target.
//	• Put a +1/+1 counter on target creature. It gains trample and haste until end of turn.
//	• Create a tapped Powerstone token. (It's an artifact with "{T}: Add {C}. This mana can't be spent to cast a nonartifact spell.")
var ArbalestEngineers = newArbalestEngineers

func newArbalestEngineers() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Arbalest Engineers",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
				cost.G,
			}),
			Colors:    []color.Color{color.Green, color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Artificer},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
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
								Text: "This creature deals 1 damage to any target.",
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "any target",
										Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
									},
								},
								Sequence: []game.Instruction{
									{
										Primitive: game.Damage{
											Amount:       game.Fixed(1),
											Recipient:    game.AnyTargetDamageRecipient(0),
											DamageSource: opt.Val(game.SourcePermanentReference()),
										},
									},
								},
							},
							game.Mode{
								Text: "Put a +1/+1 counter on target creature. It gains trample and haste until end of turn.",
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
									{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Object:      game.TargetPermanentReference(0),
											CounterKind: counter.PlusOnePlusOne,
										},
									},
									{
										Primitive: game.ApplyContinuous{
											Object: opt.Val(game.TargetPermanentReference(0)),
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Layer: game.LayerAbility,
													AddKeywords: []game.Keyword{
														game.Trample,
														game.Haste,
													},
												},
											},
											Duration: game.DurationUntilEndOfTurn,
										},
									},
								},
							},
							game.Mode{
								Text: "Create a tapped Powerstone token. (It's an artifact with \"{T}: Add {C}. This mana can't be spent to cast a nonartifact spell.\")",
								Sequence: []game.Instruction{
									{
										Primitive: game.CreateToken{
											Amount:      game.Fixed(1),
											Source:      game.TokenDef(arbalestEngineersToken),
											EntryTapped: true,
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
			When this creature enters, choose one —
			• This creature deals 1 damage to any target.
			• Put a +1/+1 counter on target creature. It gains trample and haste until end of turn.
			• Create a tapped Powerstone token. (It's an artifact with "{T}: Add {C}. This mana can't be spent to cast a nonartifact spell.")
		`,
		},
	}
}

var arbalestEngineersToken = newArbalestEngineersToken()

func newArbalestEngineersToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Powerstone",
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Powerstone},
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					AdditionalCosts: cost.Tap,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddMana{
									Amount:    game.Fixed(1),
									ManaColor: mana.C,
									SpendRider: opt.Val(game.ManaSpendRider{
										Condition:   game.ManaSpendCastArtifactSpell,
										Restriction: game.ManaSpendRestrictedToCondition,
									}),
								},
							},
						},
					}.Ability(),
				},
			},
		},
	}
}
