package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// IrenicusSVileDuplication is the card definition for Irenicus's Vile Duplication.
//
// Type: Sorcery
// Cost: {3}{U}
//
// Oracle text:
//
//	Create a token that's a copy of target creature you control, except the token has flying and it isn't legendary.
var IrenicusSVileDuplication = newIrenicusSVileDuplication

func newIrenicusSVileDuplication() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Irenicus's Vile Duplication",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature you control",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(1),
							Source: game.TokenCopyOf(game.TokenCopySpec{
								Source:          game.TokenCopySourceObject,
								Object:          game.TargetPermanentReference(0),
								SetNotLegendary: true,
								AddKeywords:     []game.Keyword{game.Flying},
							}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create a token that's a copy of target creature you control, except the token has flying and it isn't legendary.
		`,
		},
	}
}
