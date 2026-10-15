package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RetrofittedTransmogrant is the card definition for Retrofitted Transmogrant.
//
// Type: Artifact Creature — Zombie
// Cost: {B}
//
// Oracle text:
//
//	{3}{B}: Return this card from your graveyard to the battlefield tapped with two +1/+1 counters on it.
var RetrofittedTransmogrant = newRetrofittedTransmogrant

func newRetrofittedTransmogrant() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Retrofitted Transmogrant",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Zombie},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{3}{B}: Return this card from your graveyard to the battlefield tapped with two +1/+1 counters on it.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(3),
						cost.B,
					}),
					ZoneOfFunction: zone.Graveyard,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.PutOnBattlefield{
											Source:      game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceSource}),
											EntryTapped: true,
											EntryCounters: []game.CounterPlacement{
												game.CounterPlacement{
													Kind:   counter.PlusOnePlusOne,
													Amount: 2,
												},
											},
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
			{3}{B}: Return this card from your graveyard to the battlefield tapped with two +1/+1 counters on it.
		`,
		},
	}
}
