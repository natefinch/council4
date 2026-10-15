package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MentalMisstep is the card definition for Mental Misstep.
//
// Type: Instant
// Cost: {U/P}
//
// Oracle text:
//
//	({U/P} can be paid with either {U} or 2 life.)
//	Counter target spell with mana value 1.
var MentalMisstep = newMentalMisstep

func newMentalMisstep() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Mental Misstep",
			ManaCost: opt.Val(cost.Mana{
				cost.PhyrexianMana(mana.U),
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target spell with mana value 1",
						Allow:      game.TargetAllowStackObject,
						Predicate: game.TargetPredicate{
							StackObjectKinds: []game.StackObjectKind{game.StackSpell},
							ManaValue:        opt.Val(compare.Int{Op: compare.Equal, Value: 1}),
						},
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.CounterObject{
							Object: game.TargetStackObjectReference(0),
						},
					},
				},
			}.Ability()),
			OracleText: `
			({U/P} can be paid with either {U} or 2 life.)
			Counter target spell with mana value 1.
		`,
		},
	}
}
