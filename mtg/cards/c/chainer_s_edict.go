package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ChainerSEdict is the card definition for Chainer's Edict.
//
// Type: Sorcery
// Cost: {1}{B}
//
// Oracle text:
//
//	Target player sacrifices a creature of their choice.
//	Flashback {5}{B}{B} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
var ChainerSEdict = newChainerSEdict

func newChainerSEdict() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Chainer's Edict",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Flashback {5}{B}{B}",
					KeywordAbilities: []game.KeywordAbility{
						game.FlashbackKeyword{
							Cost: cost.Mana{
								cost.O(5),
								cost.B,
								cost.B,
							},
						},
					},
				},
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
						Primitive: game.SacrificePermanents{
							Player:    game.TargetPlayerReference(0),
							Amount:    game.Fixed(1),
							Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
				},
			}.Ability()),
			OracleText: `
			Target player sacrifices a creature of their choice.
			Flashback {5}{B}{B} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
		`,
		},
	}
}
