package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// PearlLakeAncient is the card definition for Pearl Lake Ancient.
//
// Type: Creature — Leviathan
// Cost: {5}{U}{U}
//
// Oracle text:
//
//	Flash
//	This spell can't be countered.
//	Prowess (Whenever you cast a noncreature spell, this creature gets +1/+1 until end of turn.)
//	Return three lands you control to their owner's hand: Return this creature to its owner's hand.
var PearlLakeAncient = newPearlLakeAncient

func newPearlLakeAncient() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Pearl Lake Ancient",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.U,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Leviathan},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 7}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
				game.CantBeCounteredStaticBody,
				game.ProwessStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Return three lands you control to their owner's hand: Return this creature to its owner's hand.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:               cost.AdditionalReturnToHand,
							Text:               "Return three lands you control to their owner's hand",
							Amount:             3,
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
										Primitive: game.MovePermanent{
											Object:      game.SourcePermanentReference(),
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
			OracleText: `
			Flash
			This spell can't be countered.
			Prowess (Whenever you cast a noncreature spell, this creature gets +1/+1 until end of turn.)
			Return three lands you control to their owner's hand: Return this creature to its owner's hand.
		`,
		},
	}
}
