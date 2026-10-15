package b

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

// BafflingEnd is the card definition for Baffling End.
//
// Type: Enchantment
// Cost: {1}{W}
//
// Oracle text:
//
//	When this enchantment enters, exile target creature an opponent controls with mana value 3 or less.
//	When this enchantment leaves the battlefield, target opponent creates a 3/3 green Dinosaur creature token with trample.
var BafflingEnd = newBafflingEnd

func newBafflingEnd() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Baffling End",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Enchantment},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this enchantment enters, exile target creature an opponent controls with mana value 3 or less.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentEnteredBattlefield,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target creature an opponent controls with mana value 3 or less",
										Allow:      game.TargetAllowPermanent,
										Selection: opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerOpponent, ManaValue: opt.Val(compare.Int{
											Op:    compare.LessOrEqual,
											Value: 3,
										})}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MovePermanent{
											Object:      game.TargetPermanentReference(0),
											Destination: zone.Exile,
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.TriggeredAbility{
					Text: "When this enchantment leaves the battlefield, target opponent creates a 3/3 green Dinosaur creature token with trample.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventZoneChanged,
							Source:          game.TriggerSourceSelf,
							MatchFromZone:   true,
							FromZone:        zone.Battlefield,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target opponent",
										Allow:      game.TargetAllowPlayer,
										Selection:  opt.Val(game.Selection{Player: game.PlayerOpponent}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount:    game.Fixed(1),
											Source:    game.TokenDef(bafflingEndToken),
											Recipient: opt.Val(game.TargetPlayerReference(0)),
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
			When this enchantment enters, exile target creature an opponent controls with mana value 3 or less.
			When this enchantment leaves the battlefield, target opponent creates a 3/3 green Dinosaur creature token with trample.
		`,
		},
	}
}

var bafflingEndToken = newBafflingEndToken()

func newBafflingEndToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Dinosaur",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dinosaur},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
			},
		},
	}
}
