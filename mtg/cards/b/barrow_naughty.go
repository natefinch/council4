package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// BarrowNaughty is the card definition for Barrow Naughty.
//
// Type: Creature — Faerie
// Cost: {1}{B}
//
// Oracle text:
//
//	Flying
//	This creature has lifelink as long as you control another Faerie.
//	{2}{B}: This creature gets +1/+0 until end of turn.
var BarrowNaughty = newBarrowNaughty

func newBarrowNaughty() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Barrow Naughty",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Faerie},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.StaticAbility{
					Text: "This creature has lifelink as long as you control another Faerie.",
					Condition: opt.Val(game.Condition{
						Text: "as long as you control another Faerie",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{SubtypesAny: []types.Sub{types.Faerie}, ExcludeSource: true},
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerAbility,
							AddKeywords:    []game.Keyword{game.Lifelink},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{2}{B}: This creature gets +1/+0 until end of turn.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(2),
						cost.B,
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ModifyPT{
											Object:     game.SourcePermanentReference(),
											PowerDelta: game.Fixed(1),
											Duration:   game.DurationUntilEndOfTurn,
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
			This creature has lifelink as long as you control another Faerie.
			{2}{B}: This creature gets +1/+0 until end of turn.
		`,
		},
	}
}
