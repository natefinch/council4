package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// LightningDragon is the card definition for Lightning Dragon.
//
// Type: Creature — Dragon
// Cost: {2}{R}{R}
//
// Oracle text:
//
//	Flying
//	Echo {2}{R}{R} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	{R}: This creature gets +1/+0 until end of turn.
var LightningDragon = newLightningDragon

func newLightningDragon() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Lightning Dragon",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dragon},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{R}: This creature gets +1/+0 until end of turn.",
					ManaCost:       opt.Val(cost.Mana{cost.R}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ModifyPT{
									Object:         game.SourcePermanentReference(),
									PowerDelta:     game.Fixed(1),
									ToughnessDelta: game.Fixed(0),
									Duration:       game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.EchoTriggeredAbility(cost.Mana{cost.O(2), cost.R, cost.R}),
			},
			OracleText: `
			Flying
			Echo {2}{R}{R} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
			{R}: This creature gets +1/+0 until end of turn.
		`,
		},
	}
}
