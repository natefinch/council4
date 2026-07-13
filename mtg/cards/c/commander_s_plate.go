package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CommanderSPlate is the card definition for Commander's Plate.
//
// Type: Artifact — Equipment
// Cost: {1}
//
// Oracle text:
//
//	Equipped creature gets +3/+3 and has protection from each color that's not in your commander's color identity.
//	Equip commander {3}
//	Equip {5}
var CommanderSPlate = newCommanderSPlate

func newCommanderSPlate() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Commander's Plate",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							PowerDelta:     3,
							ToughnessDelta: 3,
						},
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
							AddAbilities: []game.Ability{
								new(game.ProtectionFromNonCommanderIdentityColorsStaticAbility()),
							},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipCommanderActivatedAbility(cost.Mana{cost.O(3)}),
				game.EquipActivatedAbility(cost.Mana{cost.O(5)}),
			},
			OracleText: `
			Equipped creature gets +3/+3 and has protection from each color that's not in your commander's color identity.
			Equip commander {3}
			Equip {5}
		`,
		},
	}
}
