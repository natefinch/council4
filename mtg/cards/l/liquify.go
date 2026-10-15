package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Liquify is the card definition for Liquify.
//
// Type: Instant
// Cost: {2}{U}
//
// Oracle text:
//
//	Counter target spell with mana value 3 or less. If that spell is countered this way, exile it instead of putting it into its owner's graveyard.
var Liquify = newLiquify

func newLiquify() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Liquify",
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
						Constraint: "target spell with mana value 3 or less",
						Allow:      game.TargetAllowStackObject,
						Predicate: game.TargetPredicate{
							StackObjectKinds: []game.StackObjectKind{game.StackSpell},
							ManaValue:        opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 3}),
						},
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.CounterObject{
							Object:       game.TargetStackObjectReference(0),
							ExileInstead: true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Counter target spell with mana value 3 or less. If that spell is countered this way, exile it instead of putting it into its owner's graveyard.
		`,
		},
	}
}
