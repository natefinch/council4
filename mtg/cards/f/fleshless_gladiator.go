package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// FleshlessGladiator is the card definition for Fleshless Gladiator.
//
// Type: Creature — Phyrexian Skeleton
// Cost: {1}{B}
//
// Oracle text:
//
//	Corrupted — {2}{B}: Return this card from your graveyard to the battlefield tapped. You lose 1 life. Activate only if an opponent has three or more poison counters.
var FleshlessGladiator = newFleshlessGladiator

func newFleshlessGladiator() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Fleshless Gladiator",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian, types.Skeleton},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Corrupted — {2}{B}: Return this card from your graveyard to the battlefield tapped. You lose 1 life. Activate only if an opponent has three or more poison counters.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(2),
						cost.B,
					}),
					ZoneOfFunction: zone.Graveyard,
					ActivationCondition: opt.Val(game.Condition{
						Text:                     "only if an opponent has three or more poison counters",
						AnyOpponentPoisonAtLeast: 3,
						SourceCounterKind:        counter.PlusOnePlusOne,
					}),
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.PutOnBattlefield{
											Source:      game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceSource}),
											EntryTapped: true,
										},
									},
									game.Instruction{
										Primitive: game.LoseLife{
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
			Corrupted — {2}{B}: Return this card from your graveyard to the battlefield tapped. You lose 1 life. Activate only if an opponent has three or more poison counters.
		`,
		},
	}
}
