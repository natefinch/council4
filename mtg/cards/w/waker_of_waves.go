package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// WakerOfWaves is the card definition for Waker of Waves.
//
// Type: Creature — Whale
// Cost: {5}{U}{U}
//
// Oracle text:
//
//	Creatures your opponents control get -1/-0.
//	{1}{U}, Discard this card: Look at the top two cards of your library. Put one of them into your hand and the other into your graveyard.
var WakerOfWaves = newWakerOfWaves

func newWakerOfWaves() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Waker of Waves",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.U,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Whale},
			Power:     opt.Val(game.PT{Value: 7}),
			Toughness: opt.Val(game.PT{Value: 7}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Creatures your opponents control get -1/-0.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:   game.DurationPermanent,
							Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerOpponent}),
							Layer:      game.LayerPowerToughnessModify,
							PowerDelta: -1,
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{1}{U}, Discard this card: Look at the top two cards of your library. Put one of them into your hand and the other into your graveyard.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(1),
						cost.U,
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalDiscard,
							Text:        "Discard this card",
							Amount:      1,
							Source:      zone.Hand,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Hand,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Dig{
											Player: game.ControllerReference(),
											Look:   game.Fixed(2),
											Take:   game.Fixed(1),
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
			Creatures your opponents control get -1/-0.
			{1}{U}, Discard this card: Look at the top two cards of your library. Put one of them into your hand and the other into your graveyard.
		`,
		},
	}
}
