package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// FrenziedArynx is the card definition for Frenzied Arynx.
//
// Type: Creature — Cat Beast
// Cost: {2}{R}{G}
//
// Oracle text:
//
//	Riot (This creature enters with your choice of a +1/+1 counter or haste.)
//	Trample
//	{4}{R}{G}: This creature gets +3/+0 until end of turn.
var FrenziedArynx = newFrenziedArynx

func newFrenziedArynx() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Frenzied Arynx",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
				cost.G,
			}),
			Colors:    []color.Color{color.Green, color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Cat, types.Beast},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.RiotStaticBody,
				game.TrampleStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{4}{R}{G}: This creature gets +3/+0 until end of turn.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(4),
						cost.R,
						cost.G,
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ModifyPT{
											Object:     game.SourcePermanentReference(),
											PowerDelta: game.Fixed(3),
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
			Riot (This creature enters with your choice of a +1/+1 counter or haste.)
			Trample
			{4}{R}{G}: This creature gets +3/+0 until end of turn.
		`,
		},
	}
}
