package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// VoiceOfTheWoods is the card definition for Voice of the Woods.
//
// Type: Creature — Elf
// Cost: {3}{G}{G}
//
// Oracle text:
//
//	Tap five untapped Elves you control: Create a 7/7 green Elemental creature token with trample. (It can deal excess combat damage to the player or planeswalker it's attacking.)
var VoiceOfTheWoods = newVoiceOfTheWoods

func newVoiceOfTheWoods() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Voice of the Woods",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elf},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Tap five untapped Elves you control: Create a 7/7 green Elemental creature token with trample. (It can deal excess combat damage to the player or planeswalker it's attacking.)",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTapPermanents,
							Text:        "Tap five untapped Elves you control",
							Amount:      5,
							SubtypesAny: cost.SubtypeSet{0: types.Elf},
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(voiceOfTheWoodsToken),
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
			Tap five untapped Elves you control: Create a 7/7 green Elemental creature token with trample. (It can deal excess combat damage to the player or planeswalker it's attacking.)
		`,
		},
	}
}

var voiceOfTheWoodsToken = newVoiceOfTheWoodsToken()

func newVoiceOfTheWoodsToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Elemental",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elemental},
			Power:     opt.Val(game.PT{Value: 7}),
			Toughness: opt.Val(game.PT{Value: 7}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
			},
		},
	}
}
