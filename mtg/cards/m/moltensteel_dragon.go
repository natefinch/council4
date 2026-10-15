package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// MoltensteelDragon is the card definition for Moltensteel Dragon.
//
// Type: Artifact Creature — Phyrexian Dragon
// Cost: {4}{R/P}{R/P}
//
// Oracle text:
//
//	({R/P} can be paid with either {R} or 2 life.)
//	Flying
//	{R/P}: This creature gets +1/+0 until end of turn.
var MoltensteelDragon = newMoltensteelDragon

func newMoltensteelDragon() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Moltensteel Dragon",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.PhyrexianMana(mana.R),
				cost.PhyrexianMana(mana.R),
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian, types.Dragon},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{R/P}: This creature gets +1/+0 until end of turn.",
					ManaCost: opt.Val(cost.Mana{
						cost.PhyrexianMana(mana.R),
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
			({R/P} can be paid with either {R} or 2 life.)
			Flying
			{R/P}: This creature gets +1/+0 until end of turn.
		`,
		},
	}
}
