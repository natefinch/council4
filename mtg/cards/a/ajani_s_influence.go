package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AjaniSInfluence is the card definition for Ajani's Influence.
//
// Type: Sorcery
// Cost: {2}{W}{W}
//
// Oracle text:
//
//	Put two +1/+1 counters on target creature.
//	Look at the top five cards of your library. You may reveal a white card from among them and put it into your hand. Put the rest on the bottom of your library in a random order.
var AjaniSInfluence = newAjaniSInfluence()

func newAjaniSInfluence() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Ajani's Influence",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.AddCounter{
							Amount:      game.Fixed(2),
							Object:      game.TargetPermanentReference(0),
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					{
						Primitive: game.Dig{
							Player:    game.ControllerReference(),
							Look:      game.Fixed(5),
							Take:      game.Fixed(1),
							Remainder: game.DigRemainderLibraryBottom,
							Filter:    opt.Val(game.Selection{ColorsAny: []color.Color{color.White}}),
							TakeUpTo:  true,
							Reveal:    true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Put two +1/+1 counters on target creature.
			Look at the top five cards of your library. You may reveal a white card from among them and put it into your hand. Put the rest on the bottom of your library in a random order.
		`,
		},
	}
}
