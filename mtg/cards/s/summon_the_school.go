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

// SummonTheSchool is the card definition for Summon the School.
//
// Type: Kindred Sorcery — Merfolk
// Cost: {3}{W}
//
// Oracle text:
//
//	Create two 1/1 blue Merfolk Wizard creature tokens.
//	Tap four untapped Merfolk you control: Return this card from your graveyard to your hand.
var SummonTheSchool = newSummonTheSchool

func newSummonTheSchool() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Summon the School",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors:   []color.Color{color.White},
			Types:    []types.Card{types.Kindred, types.Sorcery},
			Subtypes: []types.Sub{types.Merfolk},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Tap four untapped Merfolk you control: Return this card from your graveyard to your hand.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTapPermanents,
							Text:        "Tap four untapped Merfolk you control",
							Amount:      4,
							SubtypesAny: cost.SubtypeSet{0: types.Merfolk},
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Graveyard,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MoveCard{
											Card: game.CardReference{
												Kind: game.CardReferenceSource,
											},
											FromZone:    zone.Graveyard,
											Destination: zone.Hand,
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
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(2),
							Source: game.TokenDef(summonTheSchoolToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create two 1/1 blue Merfolk Wizard creature tokens.
			Tap four untapped Merfolk you control: Return this card from your graveyard to your hand.
		`,
		},
	}
}

var summonTheSchoolToken = newSummonTheSchoolToken()

func newSummonTheSchoolToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Merfolk Wizard",
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Merfolk, types.Wizard},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
