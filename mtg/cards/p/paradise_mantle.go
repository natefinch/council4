package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ParadiseMantle is the card definition for Paradise Mantle.
//
// Type: Artifact — Equipment
// Cost: {0}
//
// Oracle text:
//
//	Equipped creature has "{T}: Add one mana of any color."
//	Equip {1}
var ParadiseMantle = newParadiseMantle

func newParadiseMantle() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Paradise Mantle",
			ManaCost: opt.Val(cost.Mana{
				cost.O(0),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Equipped creature has \"{T}: Add one mana of any color.\"",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								new(game.TapManaChoiceAbility(mana.W, mana.U, mana.B, mana.R, mana.G)),
							},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(1)}),
			},
			OracleText: `
			Equipped creature has "{T}: Add one mana of any color."
			Equip {1}
		`,
		},
	}
}
