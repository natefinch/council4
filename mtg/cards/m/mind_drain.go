package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// MindDrain is the card definition for Mind Drain.
//
// Type: Sorcery
// Cost: {2}{B}
//
// Oracle text:
//
//	Target opponent discards two cards, mills a card, and loses 1 life. You gain 1 life. (To mill a card, a player puts the top card of their library into their graveyard.)
var MindDrain = newMindDrain

func newMindDrain() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Mind Drain",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
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
						Primitive: game.Discard{
							Amount: game.Fixed(2),
							Player: game.TargetPlayerReference(0),
						},
					},
					{
						Primitive: game.MoveTopOfLibrary{
							Amount:      game.Fixed(1),
							Player:      game.TargetPlayerReference(0),
							Destination: zone.Graveyard,
						},
					},
					{
						Primitive: game.LoseLife{
							Amount: game.Fixed(1),
							Player: game.TargetPlayerReference(0),
						},
					},
					{
						Primitive: game.GainLife{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Target opponent discards two cards, mills a card, and loses 1 life. You gain 1 life. (To mill a card, a player puts the top card of their library into their graveyard.)
		`,
		},
	}
}
