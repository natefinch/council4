package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// HenchfiendOfUkor is the card definition for Henchfiend of Ukor.
//
// Type: Creature — Ogre
// Cost: {3}{R}
//
// Oracle text:
//
//	Haste
//	Echo {1}{B} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	{B/R}: This creature gets +1/+0 until end of turn.
var HenchfiendOfUkor = newHenchfiendOfUkor

func newHenchfiendOfUkor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Henchfiend of Ukor",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Ogre},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.HasteStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{B/R}: This creature gets +1/+0 until end of turn.",
					ManaCost:       opt.Val(cost.Mana{cost.HybridMana(mana.B, mana.R)}),
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
				game.EchoTriggeredAbility(cost.Mana{cost.O(1), cost.B}),
			},
			OracleText: `
			Haste
			Echo {1}{B} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
			{B/R}: This creature gets +1/+0 until end of turn.
		`,
		},
	}
}
