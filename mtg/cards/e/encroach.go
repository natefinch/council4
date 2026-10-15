package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Encroach is the card definition for Encroach.
//
// Type: Sorcery
// Cost: {B}
//
// Oracle text:
//
//	Target player reveals their hand. You choose a nonbasic land card from it. That player discards that card.
var Encroach = newEncroach

func newEncroach() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Encroach",
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
						Constraint: "Target player",
						Allow:      game.TargetAllowPlayer,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.ChooseDiscardFromHand{
							Player:    game.TargetPlayerReference(0),
							Selection: game.Selection{RequiredTypesAny: []types.Card{types.Land}, ExcludedSupertype: types.Basic},
						},
					},
				},
			}.Ability()),
			OracleText: `
			Target player reveals their hand. You choose a nonbasic land card from it. That player discards that card.
		`,
		},
	}
}
