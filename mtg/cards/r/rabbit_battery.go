package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RabbitBattery is the card definition for Rabbit Battery.
//
// Type: Artifact Creature — Equipment Rabbit
// Cost: {R}
//
// Oracle text:
//
//	Haste
//	Equipped creature gets +1/+1 and has haste.
//	Reconfigure {R} ({R}: Attach to target creature you control; or unattach from a creature. Reconfigure only as a sorcery. While attached, this isn't a creature.)
var RabbitBattery = newRabbitBattery

func newRabbitBattery() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Rabbit Battery",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Equipment, types.Rabbit},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.HasteStaticBody,
				game.StaticAbility{
					Text: "Equipped creature gets +1/+1 and has haste.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Haste},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ReconfigureActivatedAbility(cost.Mana{cost.R}),
			},
			OracleText: `
			Haste
			Equipped creature gets +1/+1 and has haste.
			Reconfigure {R} ({R}: Attach to target creature you control; or unattach from a creature. Reconfigure only as a sorcery. While attached, this isn't a creature.)
		`,
		},
	}
}
