package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DregsOfSorrow is the card definition for Dregs of Sorrow.
//
// Type: Sorcery
// Cost: {X}{4}{B}
//
// Oracle text:
//
//	Destroy X target nonblack creatures. Draw X cards.
var DregsOfSorrow = newDregsOfSorrow

func newDregsOfSorrow() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Dregs of Sorrow",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.O(4),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets:   0,
						MaxTargets:   20,
						Constraint:   "x target nonblack creatures",
						Allow:        game.TargetAllowPermanent,
						Selection:    opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, ExcludedColors: []color.Color{color.Black}}),
						CountEqualsX: true,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Object: game.AllTargetPermanentsReference(0),
						},
					},
					{
						Primitive: game.Draw{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind: game.DynamicAmountX,
							}),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy X target nonblack creatures. Draw X cards.
		`,
		},
	}
}
