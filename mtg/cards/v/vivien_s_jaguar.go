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

// VivienSJaguar is the card definition for Vivien's Jaguar.
//
// Type: Creature — Cat Spirit
// Cost: {2}{G}
//
// Oracle text:
//
//	Reach (This creature can block creatures with flying.)
//	{2}{G}: Return this card from your graveyard to your hand. Activate only if you control a Vivien planeswalker.
var VivienSJaguar = newVivienSJaguar

func newVivienSJaguar() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Vivien's Jaguar",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Cat, types.Spirit},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.ReachStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{2}{G}: Return this card from your graveyard to your hand. Activate only if you control a Vivien planeswalker.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(2),
						cost.G,
					}),
					ZoneOfFunction: zone.Graveyard,
					ActivationCondition: opt.Val(game.Condition{
						Text: "only if you control a Vivien planeswalker",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{RequiredTypes: []types.Card{types.Planeswalker}, SubtypesAny: []types.Sub{types.Vivien}},
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
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
			OracleText: `
			Reach (This creature can block creatures with flying.)
			{2}{G}: Return this card from your graveyard to your hand. Activate only if you control a Vivien planeswalker.
		`,
		},
	}
}
