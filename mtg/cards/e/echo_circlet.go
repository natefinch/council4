package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// EchoCirclet is the card definition for Echo Circlet.
//
// Type: Artifact — Equipment
// Cost: {2}
//
// Oracle text:
//
//	Equipped creature can block an additional creature each combat.
//	Equip {1}
var EchoCirclet = newEchoCirclet()

func newEchoCirclet() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Echo Circlet",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:                 game.RuleEffectCanBlockAdditional,
							AffectedAttached:     true,
							AdditionalBlockCount: 1,
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(1)}),
			},
			OracleText: `
			Equipped creature can block an additional creature each combat.
			Equip {1}
		`,
		},
	}
}
