package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// MaskwoodNexus is the card definition for Maskwood Nexus.
//
// Type: Artifact
// Cost: {4}
//
// Oracle text:
//
//	Creatures you control are every creature type. The same is true for creature spells you control and creature cards you own that aren't on the battlefield.
//	{3}, {T}: Create a 2/2 blue Shapeshifter creature token with changeling. (It is every creature type.)
var MaskwoodNexus = newMaskwoodNexus

func newMaskwoodNexus() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Maskwood Nexus",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types: []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Creatures you control are every creature type. The same is true for creature spells you control and creature cards you own that aren't on the battlefield.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:             game.DurationPermanent,
							Group:                game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							Layer:                game.LayerType,
							SpellSelection:       &game.Selection{RequiredTypes: []types.Card{types.Creature}},
							AddEveryCreatureType: true,
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{3}, {T}: Create a 2/2 blue Shapeshifter creature token with changeling. (It is every creature type.)",
					ManaCost: opt.Val(cost.Mana{
						cost.O(3),
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
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
											Source: game.TokenDef(maskwoodNexusToken),
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
			Creatures you control are every creature type. The same is true for creature spells you control and creature cards you own that aren't on the battlefield.
			{3}, {T}: Create a 2/2 blue Shapeshifter creature token with changeling. (It is every creature type.)
		`,
		},
	}
}

var maskwoodNexusToken = newMaskwoodNexusToken()

func newMaskwoodNexusToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Shapeshifter",
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Shapeshifter},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.ChangelingStaticBody,
			},
		},
	}
}
