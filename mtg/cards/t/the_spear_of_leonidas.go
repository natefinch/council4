package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TheSpearOfLeonidas is the card definition for The Spear of Leonidas.
//
// Type: Legendary Artifact — Equipment
// Cost: {2}{R}
//
// Oracle text:
//
//	Whenever equipped creature attacks, choose one —
//	• Bull Rush — It gains double strike until end of turn.
//	• Summon — Create Phobos, a legendary 3/2 red Horse creature token.
//	• Revelation — Discard two cards, then draw two cards.
//	Equip {2}
var TheSpearOfLeonidas = newTheSpearOfLeonidas

func newTheSpearOfLeonidas() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "The Spear of Leonidas",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors:     []color.Color{color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact},
			Subtypes:   []types.Sub{types.Equipment},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(2)}),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventAttackerDeclared,
							Source:           game.TriggerSourceAttachedPermanent,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Text: "Bull Rush — It gains double strike until end of turn.",
								Sequence: []game.Instruction{
									{
										Primitive: game.ApplyContinuous{
											Object: opt.Val(game.EventPermanentReference()),
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Layer: game.LayerAbility,
													AddKeywords: []game.Keyword{
														game.DoubleStrike,
													},
												},
											},
											Duration: game.DurationUntilEndOfTurn,
										},
									},
								},
							},
							game.Mode{
								Text: "Summon — Create Phobos, a legendary 3/2 red Horse creature token.",
								Sequence: []game.Instruction{
									{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(theSpearOfLeonidasToken),
										},
									},
								},
							},
							game.Mode{
								Text: "Revelation — Discard two cards, then draw two cards.",
								Sequence: []game.Instruction{
									{
										Primitive: game.Discard{
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
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			OracleText: `
			Whenever equipped creature attacks, choose one —
			• Bull Rush — It gains double strike until end of turn.
			• Summon — Create Phobos, a legendary 3/2 red Horse creature token.
			• Revelation — Discard two cards, then draw two cards.
			Equip {2}
		`,
		},
	}
}

var theSpearOfLeonidasToken = newTheSpearOfLeonidasToken()

func newTheSpearOfLeonidasToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:       "Phobos",
			Colors:     []color.Color{color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Horse},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 2}),
		},
	}
}
