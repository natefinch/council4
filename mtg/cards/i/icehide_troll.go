package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// IcehideTroll is the card definition for Icehide Troll.
//
// Type: Snow Creature — Troll Warrior
// Cost: {2}{G}
//
// Oracle text:
//
//	{S}{S}: This creature gets +2/+0 and gains indestructible until end of turn. Tap it. (Damage and effects that say "destroy" don't destroy it. {S} can be paid with one mana from a snow source.)
var IcehideTroll = newIcehideTroll

func newIcehideTroll() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Icehide Troll",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:     []color.Color{color.Green},
			Supertypes: []types.Super{types.Snow},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Troll, types.Warrior},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{S}{S}: This creature gets +2/+0 and gains indestructible until end of turn. Tap it. (Damage and effects that say \"destroy\" don't destroy it. {S} can be paid with one mana from a snow source.)",
					ManaCost: opt.Val(cost.Mana{
						cost.S,
						cost.S,
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ModifyPT{
											Object:     game.SourcePermanentReference(),
											PowerDelta: game.Fixed(2),
											Duration:   game.DurationUntilEndOfTurn,
										},
									},
									game.Instruction{
										Primitive: game.ApplyContinuous{
											Object: opt.Val(game.SourcePermanentReference()),
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Duration:    game.DurationPermanent,
													Layer:       game.LayerAbility,
													AddKeywords: []game.Keyword{game.Indestructible},
												},
											},
											Duration: game.DurationUntilEndOfTurn,
										},
									},
									game.Instruction{
										Primitive: game.Tap{
											Object: game.SourcePermanentReference(),
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
			{S}{S}: This creature gets +2/+0 and gains indestructible until end of turn. Tap it. (Damage and effects that say "destroy" don't destroy it. {S} can be paid with one mana from a snow source.)
		`,
		},
	}
}
