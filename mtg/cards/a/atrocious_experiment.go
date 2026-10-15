package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// AtrociousExperiment is the card definition for Atrocious Experiment.
//
// Type: Sorcery
// Cost: {2}{B}
//
// Oracle text:
//
//	Target player mills two cards, draws two cards, and loses 2 life. (To mill a card, a player puts the top card of their library into their graveyard.)
var AtrociousExperiment = newAtrociousExperiment

func newAtrociousExperiment() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Atrocious Experiment",
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
						Constraint: "Target player",
						Allow:      game.TargetAllowPlayer,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.MoveTopOfLibrary{
							Amount:      game.Fixed(2),
							Player:      game.TargetPlayerReference(0),
							Destination: zone.Graveyard,
						},
					},
					{
						Primitive: game.Draw{
							Amount: game.Fixed(2),
							Player: game.TargetPlayerReference(0),
						},
					},
					{
						Primitive: game.LoseLife{
							Amount: game.Fixed(2),
							Player: game.TargetPlayerReference(0),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Target player mills two cards, draws two cards, and loses 2 life. (To mill a card, a player puts the top card of their library into their graveyard.)
		`,
		},
	}
}
