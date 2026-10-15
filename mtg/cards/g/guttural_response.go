package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GutturalResponse is the card definition for Guttural Response.
//
// Type: Instant
// Cost: {R/G}
//
// Oracle text:
//
//	Counter target blue instant spell.
var GutturalResponse = newGutturalResponse

func newGutturalResponse() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Guttural Response",
			ManaCost: opt.Val(cost.Mana{
				cost.HybridMana(mana.R, mana.G),
			}),
			Colors: []color.Color{color.Green, color.Red},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target blue instant spell",
						Allow:      game.TargetAllowStackObject,
						Predicate: game.TargetPredicate{
							SpellCardTypes:   []types.Card{types.Instant},
							StackObjectKinds: []game.StackObjectKind{game.StackSpell},
							SpellColors:      []color.Color{color.Blue},
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
			Counter target blue instant spell.
		`,
		},
	}
}
