package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ManaforceMace is the card definition for Manaforce Mace.
//
// Type: Artifact — Equipment
// Cost: {4}
//
// Oracle text:
//
//	Domain — Equipped creature gets +1/+1 for each basic land type among lands you control.
//	Equip {3}
var ManaforceMace = newManaforceMace

func newManaforceMace() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Manaforce Mace",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Domain — Equipped creature gets +1/+1 for each basic land type among lands you control.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:    game.LayerPowerToughnessModify,
							PowerDeltaDynamic: opt.Val(game.DynamicAmount{
								Kind:        game.DynamicAmountControllerBasicLandTypeCount,
								Multiplier:  1,
								CounterKind: counter.PlusOnePlusOne,
							}),
							ToughnessDeltaDynamic: opt.Val(game.DynamicAmount{
								Kind:        game.DynamicAmountControllerBasicLandTypeCount,
								Multiplier:  1,
								CounterKind: counter.PlusOnePlusOne,
							}),
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(3)}),
			},
			OracleText: `
			Domain — Equipped creature gets +1/+1 for each basic land type among lands you control.
			Equip {3}
		`,
		},
	}
}
