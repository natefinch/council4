package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// EmpyrialPlate is the card definition for Empyrial Plate.
//
// Type: Artifact — Equipment
// Cost: {2}
//
// Oracle text:
//
//	Equipped creature gets +1/+1 for each card in your hand.
//	Equip {2}
var EmpyrialPlate = newEmpyrialPlate

func newEmpyrialPlate() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Empyrial Plate",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Equipped creature gets +1/+1 for each card in your hand.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:    game.LayerPowerToughnessModify,
							PowerDeltaDynamic: opt.Val(game.DynamicAmount{
								Kind:        game.DynamicAmountCountCardsInZone,
								Multiplier:  1,
								CounterKind: counter.PlusOnePlusOne,
								Player:      func() *game.PlayerReference { ref := game.ControllerReference(); return &ref }(),
								CardZone:    zone.Hand,
								Selection:   &game.Selection{},
							}),
							ToughnessDeltaDynamic: opt.Val(game.DynamicAmount{
								Kind:        game.DynamicAmountCountCardsInZone,
								Multiplier:  1,
								CounterKind: counter.PlusOnePlusOne,
								Player:      func() *game.PlayerReference { ref := game.ControllerReference(); return &ref }(),
								CardZone:    zone.Hand,
								Selection:   &game.Selection{},
							}),
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(2)}),
			},
			OracleText: `
			Equipped creature gets +1/+1 for each card in your hand.
			Equip {2}
		`,
		},
	}
}
