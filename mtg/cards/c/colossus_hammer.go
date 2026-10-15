package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ColossusHammer is the card definition for Colossus Hammer.
//
// Type: Artifact — Equipment
// Cost: {1}
//
// Oracle text:
//
//	Equipped creature gets +10/+10 and loses flying.
//	Equip {8} ({8}: Attach to target creature you control. Equip only as a sorcery.)
var ColossusHammer = newColossusHammer

func newColossusHammer() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Colossus Hammer",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Equipped creature gets +10/+10 and loses flying.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     10,
							ToughnessDelta: 10,
						},
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:          game.LayerAbility,
							RemoveKeywords: []game.Keyword{game.Flying},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(8)}),
			},
			OracleText: `
			Equipped creature gets +10/+10 and loses flying.
			Equip {8} ({8}: Attach to target creature you control. Equip only as a sorcery.)
		`,
		},
	}
}
