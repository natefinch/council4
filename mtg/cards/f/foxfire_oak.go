package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// FoxfireOak is the card definition for Foxfire Oak.
//
// Type: Creature — Treefolk Shaman
// Cost: {5}{G}
//
// Oracle text:
//
//	{R/G}{R/G}{R/G}: This creature gets +3/+0 until end of turn.
var FoxfireOak = newFoxfireOak

func newFoxfireOak() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Foxfire Oak",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Treefolk, types.Shaman},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 6}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{R/G}{R/G}{R/G}: This creature gets +3/+0 until end of turn.",
					ManaCost: opt.Val(cost.Mana{
						cost.HybridMana(mana.R, mana.G),
						cost.HybridMana(mana.R, mana.G),
						cost.HybridMana(mana.R, mana.G),
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
			{R/G}{R/G}{R/G}: This creature gets +3/+0 until end of turn.
		`,
		},
	}
}
