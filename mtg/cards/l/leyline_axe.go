package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LeylineAxe is the card definition for Leyline Axe.
//
// Type: Artifact — Equipment
// Cost: {4}
//
// Oracle text:
//
//	If this card is in your opening hand, you may begin the game with it on the battlefield.
//	Equipped creature gets +1/+1 and has double strike and trample.
//	Equip {3} ({3}: Attach to target creature you control. Equip only as a sorcery.)
var LeylineAxe = newLeylineAxe

func newLeylineAxe() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Leyline Axe",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					BeginsGameOnBattlefield: true,
				},
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
							AddKeywords: []game.Keyword{
								game.DoubleStrike,
								game.Trample,
							},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(3)}),
			},
			OracleText: `
			If this card is in your opening hand, you may begin the game with it on the battlefield.
			Equipped creature gets +1/+1 and has double strike and trample.
			Equip {3} ({3}: Attach to target creature you control. Equip only as a sorcery.)
		`,
		},
	}
}
