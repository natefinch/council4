package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TwinbladeSlasher is the card definition for Twinblade Slasher.
//
// Type: Creature — Elf Warrior
// Cost: {G}
//
// Oracle text:
//
//	Wither (This deals damage to creatures in the form of -1/-1 counters.)
//	{1}{G}: This creature gets +2/+2 until end of turn. Activate only once each turn.
var TwinbladeSlasher = newTwinbladeSlasher

func newTwinbladeSlasher() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Twinblade Slasher",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elf, types.Warrior},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.WitherStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{1}{G}: This creature gets +2/+2 until end of turn. Activate only once each turn.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(1),
						cost.G,
					}),
					ZoneOfFunction: zone.Battlefield,
					Timing:         game.OncePerTurn,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ModifyPT{
											Object:         game.SourcePermanentReference(),
											PowerDelta:     game.Fixed(2),
											ToughnessDelta: game.Fixed(2),
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
			Wither (This deals damage to creatures in the form of -1/-1 counters.)
			{1}{G}: This creature gets +2/+2 until end of turn. Activate only once each turn.
		`,
		},
	}
}
