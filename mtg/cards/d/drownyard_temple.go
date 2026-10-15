package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// DrownyardTemple is the card definition for Drownyard Temple.
//
// Type: Land
//
// Oracle text:
//
//	{T}: Add {C}.
//	{3}: Return this card from your graveyard to the battlefield tapped.
var DrownyardTemple = newDrownyardTemple

func newDrownyardTemple() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "Drownyard Temple",
			Types: []types.Card{types.Land},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{3}: Return this card from your graveyard to the battlefield tapped.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(3),
					}),
					ZoneOfFunction: zone.Graveyard,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.PutOnBattlefield{
											Source:      game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceSource}),
											EntryTapped: true,
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
			{3}: Return this card from your graveyard to the battlefield tapped.
		`,
		},
	}
}
