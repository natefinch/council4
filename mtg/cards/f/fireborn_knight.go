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

// FirebornKnight is the card definition for Fireborn Knight.
//
// Type: Creature — Human Knight
// Cost: {R/W}{R/W}{R/W}{R/W}
//
// Oracle text:
//
//	Double strike
//	{R/W}{R/W}{R/W}{R/W}: This creature gets +1/+1 until end of turn.
var FirebornKnight = newFirebornKnight

func newFirebornKnight() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red),
		CardFace: game.CardFace{
			Name: "Fireborn Knight",
			ManaCost: opt.Val(cost.Mana{
				cost.HybridMana(mana.R, mana.W),
				cost.HybridMana(mana.R, mana.W),
				cost.HybridMana(mana.R, mana.W),
				cost.HybridMana(mana.R, mana.W),
			}),
			Colors:    []color.Color{color.Red, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Knight},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.DoubleStrikeStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{R/W}{R/W}{R/W}{R/W}: This creature gets +1/+1 until end of turn.",
					ManaCost: opt.Val(cost.Mana{
						cost.HybridMana(mana.R, mana.W),
						cost.HybridMana(mana.R, mana.W),
						cost.HybridMana(mana.R, mana.W),
						cost.HybridMana(mana.R, mana.W),
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ModifyPT{
											Object:         game.SourcePermanentReference(),
											PowerDelta:     game.Fixed(1),
											ToughnessDelta: game.Fixed(1),
											Duration:       game.DurationUntilEndOfTurn,
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
			Double strike
			{R/W}{R/W}{R/W}{R/W}: This creature gets +1/+1 until end of turn.
		`,
		},
	}
}
