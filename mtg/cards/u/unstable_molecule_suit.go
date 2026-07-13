package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// UnstableMoleculeSuit is the card definition for Unstable Molecule Suit.
//
// Type: Artifact — Equipment
// Cost: {3}
//
// Oracle text:
//
//	Equipped creature gets +2/+2 and has indestructible.
//	Equip commander {2}
//	Equip {4}
var UnstableMoleculeSuit = newUnstableMoleculeSuit

func newUnstableMoleculeSuit() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Unstable Molecule Suit",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							PowerDelta:     2,
							ToughnessDelta: 2,
						},
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
							AddKeywords: []game.Keyword{
								game.Indestructible,
							},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipCommanderActivatedAbility(cost.Mana{cost.O(2)}),
				game.EquipActivatedAbility(cost.Mana{cost.O(4)}),
			},
			OracleText: `
			Equipped creature gets +2/+2 and has indestructible.
			Equip commander {2}
			Equip {4}
		`,
		},
	}
}
