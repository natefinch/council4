package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ExecutionerSHood is the card definition for Executioner's Hood.
//
// Type: Artifact — Equipment
// Cost: {2}
//
// Oracle text:
//
//	Equipped creature has intimidate. (This creature can't be blocked except by artifact creatures and/or creatures that share a color with it.)
//	Equip {2} ({2}: Attach to target creature you control. Equip only as a sorcery.)
var ExecutionerSHood = newExecutionerSHood

func newExecutionerSHood() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Executioner's Hood",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Equipped creature has intimidate. (This creature can't be blocked except by artifact creatures and/or creatures that share a color with it.)",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Intimidate},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(2)}),
			},
			OracleText: `
			Equipped creature has intimidate. (This creature can't be blocked except by artifact creatures and/or creatures that share a color with it.)
			Equip {2} ({2}: Attach to target creature you control. Equip only as a sorcery.)
		`,
		},
	}
}
