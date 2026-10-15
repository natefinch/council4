package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ReleaseTheGremlins is the card definition for Release the Gremlins.
//
// Type: Sorcery
// Cost: {X}{X}{R}
//
// Oracle text:
//
//	Destroy X target artifacts. Create X 2/2 red Gremlin creature tokens.
var ReleaseTheGremlins = newReleaseTheGremlins

func newReleaseTheGremlins() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Release the Gremlins",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.X,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets:   0,
						MaxTargets:   20,
						Constraint:   "x target artifacts",
						Allow:        game.TargetAllowPermanent,
						Selection:    opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact}}),
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
						Primitive: game.CreateToken{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind: game.DynamicAmountX,
							}),
							Source: game.TokenDef(releaseTheGremlinsToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy X target artifacts. Create X 2/2 red Gremlin creature tokens.
		`,
		},
	}
}

var releaseTheGremlinsToken = newReleaseTheGremlinsToken()

func newReleaseTheGremlinsToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Gremlin",
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Gremlin},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
		},
	}
}
