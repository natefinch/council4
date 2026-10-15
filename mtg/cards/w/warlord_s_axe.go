package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WarlordSAxe is the card definition for Warlord's Axe.
//
// Type: Artifact — Equipment
// Cost: {3}
//
// Oracle text:
//
//	Equipped creature gets +3/+1.
//	Equip {4} ({4}: Attach to target creature you control. Equip only as a sorcery.)
var WarlordSAxe = newWarlordSAxe

func newWarlordSAxe() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Warlord's Axe",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Equipped creature gets +3/+1.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     3,
							ToughnessDelta: 1,
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(4)}),
			},
			OracleText: `
			Equipped creature gets +3/+1.
			Equip {4} ({4}: Attach to target creature you control. Equip only as a sorcery.)
		`,
		},
	}
}
