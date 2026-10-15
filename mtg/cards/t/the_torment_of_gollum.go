package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TheTormentOfGollum is the card definition for The Torment of Gollum.
//
// Type: Sorcery
// Cost: {3}{B}
//
// Oracle text:
//
//	Target opponent reveals their hand. You choose a nonland card from it. That player discards that card.
//	Amass Orcs 2. (Put two +1/+1 counters on an Army you control. It's also an Orc. If you don't control an Army, create a 0/0 black Orc Army creature token first.)
var TheTormentOfGollum = newTheTormentOfGollum

func newTheTormentOfGollum() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "The Torment of Gollum",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
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
						},
					},
					{
						Primitive: game.Amass{
							Amount:  game.Fixed(2),
							Subtype: types.Orc,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Target opponent reveals their hand. You choose a nonland card from it. That player discards that card.
			Amass Orcs 2. (Put two +1/+1 counters on an Army you control. It's also an Orc. If you don't control an Army, create a 0/0 black Orc Army creature token first.)
		`,
		},
	}
}
