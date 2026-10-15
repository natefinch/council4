package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// EvraHalcyonWitness is the card definition for Evra, Halcyon Witness.
//
// Type: Legendary Creature — Avatar
// Cost: {4}{W}{W}
//
// Oracle text:
//
//	Lifelink
//	{4}: Exchange your life total with Evra's power.
var EvraHalcyonWitness = newEvraHalcyonWitness

func newEvraHalcyonWitness() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Evra, Halcyon Witness",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.W,
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Avatar},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.LifelinkStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{4}: Exchange your life total with Evra's power.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(4),
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ExchangeLifeTotalWithSourceCharacteristic{
											Player:         game.ControllerReference(),
											Characteristic: game.SourcePower,
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
			Lifelink
			{4}: Exchange your life total with Evra's power.
		`,
		},
	}
}
