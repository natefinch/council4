package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Counterlash is the card definition for Counterlash.
//
// Type: Instant
// Cost: {4}{U}{U}
//
// Oracle text:
//
//	Counter target spell. You may cast a spell that shares a card type with it from your hand without paying its mana cost.
var Counterlash = newCounterlash

func newCounterlash() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Counterlash",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.U,
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
							Object: game.TargetStackObjectReference(0),
						},
					},
					{
						Primitive: game.CastForFree{
							Player:    game.ControllerReference(),
							Selection: game.Selection{ExcludedTypes: []types.Card{types.Land}},
							Zone:      zone.Hand,
						},
						Optional: true,
					},
				},
			}.Ability()),
			OracleText: `
			Counter target spell. You may cast a spell that shares a card type with it from your hand without paying its mana cost.
		`,
		},
	}
}
