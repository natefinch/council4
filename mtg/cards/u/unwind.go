package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Unwind is the card definition for Unwind.
//
// Type: Instant
// Cost: {2}{U}
//
// Oracle text:
//
//	Counter target noncreature spell. Untap up to three lands.
var Unwind = newUnwind

func newUnwind() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Unwind",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target noncreature spell",
						Allow:      game.TargetAllowStackObject,
						Predicate: game.TargetPredicate{
							ExcludedSpellCardTypes: []types.Card{types.Creature},
							StackObjectKinds:       []game.StackObjectKind{game.StackSpell},
						},
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.CounterObject{
							Object: game.TargetStackObjectReference(0),
						},
					},
					{
						Primitive: game.Untap{
							Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Land}}),
							ChooseUpTo: true,
							Amount:     game.Fixed(3),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Counter target noncreature spell. Untap up to three lands.
		`,
		},
	}
}
