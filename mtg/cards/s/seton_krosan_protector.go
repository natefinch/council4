package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SetonKrosanProtector is the card definition for Seton, Krosan Protector.
//
// Type: Legendary Creature — Centaur Druid
// Cost: {G}{G}{G}
//
// Oracle text:
//
//	Tap an untapped Druid you control: Add {G}.
var SetonKrosanProtector = newSetonKrosanProtector

func newSetonKrosanProtector() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Seton, Krosan Protector",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
				cost.G,
				cost.G,
			}),
			Colors:     []color.Color{color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Centaur, types.Druid},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					Text: "Tap an untapped Druid you control: Add {G}.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTapPermanents,
							Text:        "Tap an untapped Druid you control",
							Amount:      1,
							SubtypesAny: cost.SubtypeSet{0: types.Druid},
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddMana{
											Amount:    game.Fixed(1),
											ManaColor: mana.G,
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
			Tap an untapped Druid you control: Add {G}.
		`,
		},
	}
}
