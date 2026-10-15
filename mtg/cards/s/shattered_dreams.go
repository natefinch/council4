package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ShatteredDreams is the card definition for Shattered Dreams.
//
// Type: Sorcery
// Cost: {B}
//
// Oracle text:
//
//	Target opponent reveals their hand. You choose an artifact card from it. That player discards that card.
var ShatteredDreams = newShatteredDreams

func newShatteredDreams() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Shattered Dreams",
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
							Selection: game.Selection{RequiredTypesAny: []types.Card{types.Artifact}},
						},
					},
				},
			}.Ability()),
			OracleText: `
			Target opponent reveals their hand. You choose an artifact card from it. That player discards that card.
		`,
		},
	}
}
