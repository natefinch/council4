package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PantherHabit is the card definition for Panther Habit.
//
// Type: Artifact — Equipment
// Cost: {4}
//
// Oracle text:
//
//	If equipped creature would be dealt damage, prevent that damage and put that many +1/+1 counters on it.
//	Equip {2}
var PantherHabit = newPantherHabit()

func newPantherHabit() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Panther Habit",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(2)}),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.DamagePreventionToPlusOneCountersReplacement("If equipped creature would be dealt damage, prevent that damage and put that many +1/+1 counters on it.", true, opt.V[game.Condition]{}),
			},
			OracleText: `
			If equipped creature would be dealt damage, prevent that damage and put that many +1/+1 counters on it.
			Equip {2}
		`,
		},
	}
}
