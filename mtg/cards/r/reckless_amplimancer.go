package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RecklessAmplimancer is the card definition for Reckless Amplimancer.
//
// Type: Creature — Elf Druid
// Cost: {1}{G}
//
// Oracle text:
//
//	{4}{G}: Double this creature's power and toughness until end of turn.
var RecklessAmplimancer = newRecklessAmplimancer()

func newRecklessAmplimancer() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Reckless Amplimancer",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elf, types.Druid},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{4}{G}: Double this creature's power and toughness until end of turn.",
					ManaCost:       opt.Val(cost.Mana{cost.O(4), cost.G}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									Object: opt.Val(game.SourcePermanentReference()),
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer:           game.LayerPowerToughnessModify,
											DoublePower:     true,
											DoubleToughness: true,
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{4}{G}: Double this creature's power and toughness until end of turn.
		`,
		},
	}
}
