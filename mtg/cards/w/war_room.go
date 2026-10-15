package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// WarRoom is the card definition for War Room.
//
// Type: Land
//
// Oracle text:
//
//	{T}: Add {C}.
//	{3}, {T}, Pay life equal to the number of colors in your commanders' color identity: Draw a card.
var WarRoom = newWarRoom

func newWarRoom() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "War Room",
			Types: []types.Card{types.Land},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{3}, {T}, Pay life equal to the number of colors in your commanders' color identity: Draw a card.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(3),
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
						cost.Additional{
							Kind:          cost.AdditionalPayLife,
							Text:          "Pay life equal to the number of colors in your commanders' color identity",
							AmountDynamic: cost.AdditionalDynamicCommanderColorIdentityCount,
							CounterKind:   counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
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
			{3}, {T}, Pay life equal to the number of colors in your commanders' color identity: Draw a card.
		`,
		},
	}
}
