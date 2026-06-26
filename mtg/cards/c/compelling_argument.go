package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CompellingArgument is the card definition for Compelling Argument.
//
// Type: Sorcery
// Cost: {1}{U}
//
// Oracle text:
//
//	Target player mills five cards.
//	Cycling {U} ({U}, Discard this card: Draw a card.)
var CompellingArgument = newCompellingArgument()

func newCompellingArgument() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Compelling Argument",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.U}),
			},
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
						Primitive: game.Mill{
							Amount: game.Fixed(5),
							Player: game.TargetPlayerReference(0),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Target player mills five cards.
			Cycling {U} ({U}, Discard this card: Draw a card.)
		`,
		},
	}
}
