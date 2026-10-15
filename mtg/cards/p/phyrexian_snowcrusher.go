package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// PhyrexianSnowcrusher is the card definition for Phyrexian Snowcrusher.
//
// Type: Snow Artifact Creature — Phyrexian Juggernaut
// Cost: {6}
//
// Oracle text:
//
//	This creature attacks each combat if able.
//	{1}{S}: This creature gets +1/+0 until end of turn. ({S} can be paid with one mana from a snow source.)
var PhyrexianSnowcrusher = newPhyrexianSnowcrusher

func newPhyrexianSnowcrusher() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Phyrexian Snowcrusher",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
			}),
			Supertypes: []types.Super{types.Snow},
			Types:      []types.Card{types.Artifact, types.Creature},
			Subtypes:   []types.Sub{types.Phyrexian, types.Juggernaut},
			Power:      opt.Val(game.PT{Value: 6}),
			Toughness:  opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.MustAttackStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{1}{S}: This creature gets +1/+0 until end of turn. ({S} can be paid with one mana from a snow source.)",
					ManaCost: opt.Val(cost.Mana{
						cost.O(1),
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
			This creature attacks each combat if able.
			{1}{S}: This creature gets +1/+0 until end of turn. ({S} can be paid with one mana from a snow source.)
		`,
		},
	}
}
