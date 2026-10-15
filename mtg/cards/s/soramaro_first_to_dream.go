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

// SoramaroFirstToDream is the card definition for Soramaro, First to Dream.
//
// Type: Legendary Creature — Spirit
// Cost: {4}{U}{U}
//
// Oracle text:
//
//	Flying
//	Soramaro's power and toughness are each equal to the number of cards in your hand.
//	{4}, Return a land you control to its owner's hand: Draw a card.
var SoramaroFirstToDream = newSoramaroFirstToDream

func newSoramaroFirstToDream() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Soramaro, First to Dream",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.U,
				cost.U,
			}),
			Colors:           []color.Color{color.Blue},
			Supertypes:       []types.Super{types.Legendary},
			Types:            []types.Card{types.Creature},
			Subtypes:         []types.Sub{types.Spirit},
			Power:            opt.Val(game.PT{IsStar: true}),
			Toughness:        opt.Val(game.PT{IsStar: true}),
			DynamicPower:     opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerHandSize}),
			DynamicToughness: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerHandSize}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{4}, Return a land you control to its owner's hand: Draw a card.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(4),
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:               cost.AdditionalReturnToHand,
							Text:               "Return a land you control to its owner's hand",
							Amount:             1,
							MatchPermanentType: true,
							PermanentType:      types.Land,
							CounterKind:        counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Fixed(1),
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
			Flying
			Soramaro's power and toughness are each equal to the number of cards in your hand.
			{4}, Return a land you control to its owner's hand: Draw a card.
		`,
		},
	}
}
