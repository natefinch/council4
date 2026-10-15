package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GeierReachSanitarium is the card definition for Geier Reach Sanitarium.
//
// Type: Legendary Land
//
// Oracle text:
//
//	{T}: Add {C}.
//	{2}, {T}: Each player draws a card, then discards a card.
var GeierReachSanitarium = newGeierReachSanitarium

func newGeierReachSanitarium() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:       "Geier Reach Sanitarium",
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Land},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{2}, {T}: Each player draws a card, then discards a card.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(2),
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Draw{
											Amount:      game.Fixed(1),
											PlayerGroup: game.AllPlayersReference(),
										},
									},
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
			{2}, {T}: Each player draws a card, then discards a card.
		`,
		},
	}
}
