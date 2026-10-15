package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GitaxianProbe is the card definition for Gitaxian Probe.
//
// Type: Sorcery
// Cost: {U/P}
//
// Oracle text:
//
//	({U/P} can be paid with either {U} or 2 life.)
//	Look at target player's hand.
//	Draw a card.
var GitaxianProbe = newGitaxianProbe

func newGitaxianProbe() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Gitaxian Probe",
			ManaCost: opt.Val(cost.Mana{
				cost.PhyrexianMana(mana.U),
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target player",
						Allow:      game.TargetAllowPlayer,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.LookAtHand{
							Player: game.TargetPlayerReference(0),
						},
					},
					{
						Primitive: game.Draw{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			({U/P} can be paid with either {U} or 2 life.)
			Look at target player's hand.
			Draw a card.
		`,
		},
	}
}
