package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DescentOfTheDragons is the card definition for Descent of the Dragons.
//
// Type: Sorcery
// Cost: {4}{R}{R}
//
// Oracle text:
//
//	Destroy any number of target creatures. For each creature destroyed this way, its controller creates a 4/4 red Dragon creature token with flying.
var DescentOfTheDragons = newDescentOfTheDragons

func newDescentOfTheDragons() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Descent of the Dragons",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					{
						MinTargets: 0,
						MaxTargets: 99,
						Constraint: "any number of target creatures",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.RemoveTargetsForToken{
							LinkedKey: game.LinkedKey("removed-targets-for-token"),
						},
					},
					{
						Primitive: game.CreateTokenForEachDestroyed{
							Source:    game.TokenDef(descentOfTheDragonsToken),
							LinkedKey: game.LinkedKey("removed-targets-for-token"),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy any number of target creatures. For each creature destroyed this way, its controller creates a 4/4 red Dragon creature token with flying.
		`,
		},
	}
}

var descentOfTheDragonsToken = newDescentOfTheDragonsToken()

func newDescentOfTheDragonsToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Dragon",
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dragon},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
		},
	}
}
