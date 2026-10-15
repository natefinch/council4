package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Ostracize is the card definition for Ostracize.
//
// Type: Sorcery
// Cost: {B}
//
// Oracle text:
//
//	Target opponent reveals their hand. You choose a creature card from it. That player discards that card.
var Ostracize = newOstracize

func newOstracize() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Ostracize",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "Target opponent",
						Allow:      game.TargetAllowPlayer,
						Selection:  opt.Val(game.Selection{Player: game.PlayerOpponent}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.ChooseDiscardFromHand{
							Player:    game.TargetPlayerReference(0),
							Selection: game.Selection{RequiredTypesAny: []types.Card{types.Creature}},
						},
					},
				},
			}.Ability()),
			OracleText: `
			Target opponent reveals their hand. You choose a creature card from it. That player discards that card.
		`,
		},
	}
}
