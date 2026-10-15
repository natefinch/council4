package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ImmoralBargain is the card definition for Immoral Bargain.
//
// Type: Sorcery
// Cost: {1}{B}{G}
//
// Oracle text:
//
//	As an additional cost to cast this spell, sacrifice X creatures.
//	Destroy X target nonland permanents.
var ImmoralBargain = newImmoralBargain

func newImmoralBargain() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Immoral Bargain",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
				cost.G,
			}),
			Colors: []color.Color{color.Black, color.Green},
			Types:  []types.Card{types.Sorcery},
			AdditionalCosts: []cost.Additional{
				{
					Kind:               cost.AdditionalSacrifice,
					Text:               "sacrifice X creatures",
					AmountFromX:        true,
					MatchPermanentType: true,
					PermanentType:      types.Creature,
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets:   0,
						MaxTargets:   20,
						Constraint:   "x target nonland permanents",
						Allow:        game.TargetAllowPermanent,
						Selection:    opt.Val(game.Selection{ExcludedTypes: []types.Card{types.Land}}),
						CountEqualsX: true,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Object: game.AllTargetPermanentsReference(0),
						},
					},
				},
			}.Ability()),
			OracleText: `
			As an additional cost to cast this spell, sacrifice X creatures.
			Destroy X target nonland permanents.
		`,
		},
	}
}
