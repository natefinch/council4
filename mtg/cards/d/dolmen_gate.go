package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DolmenGate is the card definition for Dolmen Gate.
//
// Type: Artifact
// Cost: {2}
//
// Oracle text:
//
//	Prevent all combat damage that would be dealt to attacking creatures you control.
var DolmenGate = newDolmenGate

func newDolmenGate() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Dolmen Gate",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types: []types.Card{types.Artifact},
			ReplacementAbilities: []game.ReplacementAbility{
				game.CombatDamagePreventionToGroupReplacement("Prevent all combat damage that would be dealt to attacking creatures you control.", game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou, CombatState: game.CombatStateAttacking}),
			},
			OracleText: `
			Prevent all combat damage that would be dealt to attacking creatures you control.
		`,
		},
	}
}
