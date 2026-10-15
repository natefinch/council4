package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LayBareTheHeart is the card definition for Lay Bare the Heart.
//
// Type: Sorcery
// Cost: {1}{B}
//
// Oracle text:
//
//	Target opponent reveals their hand. You choose a nonlegendary, nonland card from it. That player discards that card.
var LayBareTheHeart = newLayBareTheHeart

func newLayBareTheHeart() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Lay Bare the Heart",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
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
							Player:      game.TargetPlayerReference(0),
							ExcludeLand: true,
							Selection:   game.Selection{ExcludedSupertype: types.Legendary},
						},
					},
				},
			}.Ability()),
			OracleText: `
			Target opponent reveals their hand. You choose a nonlegendary, nonland card from it. That player discards that card.
		`,
		},
	}
}
