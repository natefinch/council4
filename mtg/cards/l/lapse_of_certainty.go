package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LapseOfCertainty is the card definition for Lapse of Certainty.
//
// Type: Instant
// Cost: {2}{W}
//
// Oracle text:
//
//	Counter target spell. If that spell is countered this way, put it on top of its owner's library instead of into that player's graveyard.
var LapseOfCertainty = newLapseOfCertainty()

func newLapseOfCertainty() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Lapse of Certainty",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors: []color.Color{color.White},
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
