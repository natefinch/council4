package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MemoryLapse is the card definition for Memory Lapse.
//
// Type: Instant
// Cost: {1}{U}
//
// Oracle text:
//
//	Counter target spell. If that spell is countered this way, put it on top of its owner's library instead of into that player's graveyard.
var MemoryLapse = newMemoryLapse

func newMemoryLapse() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Memory Lapse",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target spell",
						Allow:      game.TargetAllowStackObject,
						Predicate: game.TargetPredicate{
							StackObjectKinds: []game.StackObjectKind{game.StackSpell},
						},
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.CounterObject{
							Object:      game.TargetStackObjectReference(0),
							Destination: game.CounteredSpellLibraryTop,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Counter target spell. If that spell is countered this way, put it on top of its owner's library instead of into that player's graveyard.
		`,
		},
	}
}
