package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RixMaadiDungeonPalace is the card definition for Rix Maadi, Dungeon Palace.
//
// Type: Land
//
// Oracle text:
//
//	{T}: Add {C}.
//	{1}{B}{R}, {T}: Each player discards a card. Activate only as a sorcery.
var RixMaadiDungeonPalace = newRixMaadiDungeonPalace

func newRixMaadiDungeonPalace() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red),
		CardFace: game.CardFace{
			Name:  "Rix Maadi, Dungeon Palace",
			Types: []types.Card{types.Land},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{1}{B}{R}, {T}: Each player discards a card. Activate only as a sorcery.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(1),
						cost.B,
						cost.R,
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Timing:         game.SorceryOnly,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Discard{
											Amount:      game.Fixed(1),
											PlayerGroup: game.AllPlayersReference(),
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
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.C),
			},
			OracleText: `
			{T}: Add {C}.
			{1}{B}{R}, {T}: Each player discards a card. Activate only as a sorcery.
		`,
		},
	}
}
