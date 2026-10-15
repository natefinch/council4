package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// DidnTSayPlease is the card definition for Didn't Say Please.
//
// Type: Instant
// Cost: {1}{U}{U}
//
// Oracle text:
//
//	Counter target spell. Its controller mills three cards.
var DidnTSayPlease = newDidnTSayPlease

func newDidnTSayPlease() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Didn't Say Please",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
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
						Primitive: game.MoveTopOfLibrary{
							Amount:      game.Fixed(3),
							Player:      game.ObjectControllerReference(game.TargetStackObjectReference(0)),
							Destination: zone.Graveyard,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Counter target spell. Its controller mills three cards.
		`,
		},
	}
}
