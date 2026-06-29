package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SaradocMasterOfBuckland is the card definition for Saradoc, Master of Buckland.
//
// Type: Legendary Creature — Halfling Citizen
// Cost: {3}{W}
//
// Oracle text:
//
//	Whenever Saradoc or another nontoken creature you control with power 2 or less enters, create a 1/1 white Halfling creature token.
//	Tap two other untapped Halflings you control: Saradoc gets +2/+0 and gains lifelink until end of turn.
var SaradocMasterOfBuckland = newSaradocMasterOfBuckland()

func newSaradocMasterOfBuckland() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Saradoc, Master of Buckland",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Halfling, types.Citizen},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Tap two other untapped Halflings you control: Saradoc gets +2/+0 and gains lifelink until end of turn.",
					AdditionalCosts: []cost.Additional{
						{
							Kind:          cost.AdditionalTapPermanents,
							Text:          "Tap two other untapped Halflings you control",
							Amount:        2,
							ExcludeSource: true,
							SubtypesAny:   cost.SubtypeSet{types.Halfling},
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ModifyPT{
									Object:         game.SourcePermanentReference(),
									PowerDelta:     game.Fixed(2),
									ToughnessDelta: game.Fixed(0),
									Duration:       game.DurationUntilEndOfTurn,
								},
							},
							{
								Primitive: game.ApplyContinuous{
									Object: opt.Val(game.SourceCardPermanentReference()),
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											AddKeywords: []game.Keyword{
												game.Lifelink,
											},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                  game.EventPermanentEnteredBattlefield,
							Controller:             game.TriggerControllerYou,
							SubjectSelectionOrSelf: true,
							SubjectSelection:       game.Selection{RequiredTypes: []types.Card{types.Creature}, Power: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 2}), NonToken: true},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(saradocMasterOfBucklandToken),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever Saradoc or another nontoken creature you control with power 2 or less enters, create a 1/1 white Halfling creature token.
			Tap two other untapped Halflings you control: Saradoc gets +2/+0 and gains lifelink until end of turn.
		`,
		},
	}
}

var saradocMasterOfBucklandToken = newSaradocMasterOfBucklandToken()

func newSaradocMasterOfBucklandToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Halfling",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Halfling},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
