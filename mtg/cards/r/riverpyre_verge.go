package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RiverpyreVerge is the card definition for Riverpyre Verge.
//
// Type: Land
//
// Oracle text:
//
//	{T}: Add {R}.
//	{T}: Add {U}. Activate only if you control an Island or a Mountain.
var RiverpyreVerge = newRiverpyreVerge

func newRiverpyreVerge() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red),
		CardFace: game.CardFace{
			Name:  "Riverpyre Verge",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.R),
				game.ManaAbility{
					Text: "{T}: Add {U}. Activate only if you control an Island or a Mountain.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ActivationCondition: opt.Val(game.Condition{
						Text: "only if you control an Island or a Mountain",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{RequiredTypes: []types.Card{types.Land}, SubtypesAny: []types.Sub{types.Island, types.Mountain}},
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddMana{
											Amount:    game.Fixed(1),
											ManaColor: mana.U,
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
			{T}: Add {R}.
			{T}: Add {U}. Activate only if you control an Island or a Mountain.
		`,
		},
	}
}
