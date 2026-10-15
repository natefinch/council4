package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// PsychicMiasma is the card definition for Psychic Miasma.
//
// Type: Sorcery
// Cost: {1}{B}
//
// Oracle text:
//
//	Target player discards a card. If a land card is discarded this way, return Psychic Miasma to its owner's hand.
var PsychicMiasma = newPsychicMiasma

func newPsychicMiasma() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Psychic Miasma",
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
						Constraint: "Target player",
						Allow:      game.TargetAllowPlayer,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Discard{
							Amount: game.Fixed(1),
							Player: game.TargetPlayerReference(0),
						},
						PublishResult: game.ResultKey("if-you-do"),
						LocalProducts: game.LocalProducts{
							Results: []game.ResultKey{game.ResultKey("if-you-do")},
						},
					},
					{
						Primitive: game.MoveResolvingSpell{
							Destination: zone.Hand,
						},
						ResultGate: opt.Val(game.InstructionResultGate{
							Key:             "if-you-do",
							Succeeded:       game.TriTrue,
							ObjectSelection: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Land}}),
							CardOnly:        true,
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Target player discards a card. If a land card is discarded this way, return Psychic Miasma to its owner's hand.
		`,
		},
	}
}
