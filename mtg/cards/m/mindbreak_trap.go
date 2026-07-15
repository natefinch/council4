package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MindbreakTrap is the card definition for Mindbreak Trap.
//
// Type: Instant — Trap
// Cost: {2}{U}{U}
//
// Oracle text:
//
//	If an opponent cast three or more spells this turn, you may pay {0} rather than pay this spell's mana cost.
//	Exile any number of target spells.
var MindbreakTrap = newMindbreakTrap

func newMindbreakTrap() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Mindbreak Trap",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
				cost.U,
			}),
			Colors:   []color.Color{color.Blue},
			Types:    []types.Card{types.Instant},
			Subtypes: []types.Sub{types.Trap},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label:          "Pay {0}",
					ManaCost:       opt.Val(cost.Mana{cost.O(0)}),
					Condition:      cost.AlternativeConditionOpponentCastSpellsThisTurn,
					ConditionCount: 3,
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 0,
						MaxTargets: 99,
						Constraint: "any number of target spells",
						Allow:      game.TargetAllowStackObject,
						Predicate: game.TargetPredicate{
							StackObjectKinds: []game.StackObjectKind{game.StackSpell},
						},
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.ExileTargetSpells{
							Object: game.AllTargetStackObjectsReference(0),
						},
					},
				},
			}.Ability()),
			OracleText: `
			If an opponent cast three or more spells this turn, you may pay {0} rather than pay this spell's mana cost.
			Exile any number of target spells.
		`,
		},
	}
}
