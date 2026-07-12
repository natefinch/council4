package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Mockingbird is the card definition for Mockingbird.
//
// Type: Creature — Bird Bard
// Cost: {X}{U}
//
// Oracle text:
//
//	Flying
//	You may have this creature enter as a copy of any creature on the battlefield with mana value less than or equal to the amount of mana spent to cast this creature, except it's a Bird in addition to its other types and it has flying.
var Mockingbird = newMockingbird

func newMockingbird() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Mockingbird",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Bird, types.Bard},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersAsCopyWithManaSpentBound(game.EntersAsCopyReplacement("You may have this creature enter as a copy of any creature on the battlefield with mana value less than or equal to the amount of mana spent to cast this creature, except it's a Bird in addition to its other types and it has flying.", &game.Selection{RequiredTypes: []types.Card{types.Creature}}, true, false, nil, false, []game.Keyword{game.Flying}, []types.Sub{types.Bird})),
			},
			OracleText: `
			Flying
			You may have this creature enter as a copy of any creature on the battlefield with mana value less than or equal to the amount of mana spent to cast this creature, except it's a Bird in addition to its other types and it has flying.
		`,
		},
	}
}
