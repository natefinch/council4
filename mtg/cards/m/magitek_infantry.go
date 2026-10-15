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

// MagitekInfantry is the card definition for Magitek Infantry.
//
// Type: Artifact Creature — Robot Soldier
// Cost: {W}
//
// Oracle text:
//
//	This creature gets +1/+0 as long as you control another artifact.
//	{2}{W}: Search your library for a card named Magitek Infantry, put it onto the battlefield tapped, then shuffle.
var MagitekInfantry = newMagitekInfantry

func newMagitekInfantry() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Magitek Infantry",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Robot, types.Soldier},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This creature gets +1/+0 as long as you control another artifact.",
					Condition: opt.Val(game.Condition{
						Text: "as long as you control another artifact",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{RequiredTypes: []types.Card{types.Artifact}, ExcludeSource: true},
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     1,
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{2}{W}: Search your library for a card named Magitek Infantry, put it onto the battlefield tapped, then shuffle.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(2),
						cost.W,
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Search{
											Player: game.ControllerReference(),
											Spec: game.SearchSpec{
												SourceZone:   zone.Library,
												Destination:  zone.Battlefield,
												EntersTapped: true,
												Name:         "Magitek Infantry",
											},
											Amount: game.Fixed(1),
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
			This creature gets +1/+0 as long as you control another artifact.
			{2}{W}: Search your library for a card named Magitek Infantry, put it onto the battlefield tapped, then shuffle.
		`,
		},
	}
}
