package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// OrderedMigration is the card definition for Ordered Migration.
//
// Type: Sorcery
// Cost: {3}{W}{U}
//
// Oracle text:
//
//	Domain — Create a 1/1 blue Bird creature token with flying for each basic land type among lands you control.
var OrderedMigration = newOrderedMigration

func newOrderedMigration() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name: "Ordered Migration",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
				cost.U,
			}),
			Colors: []color.Color{color.Blue, color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountControllerBasicLandTypeCount,
								Multiplier: 1,
							}),
							Source: game.TokenDef(orderedMigrationToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Domain — Create a 1/1 blue Bird creature token with flying for each basic land type among lands you control.
		`,
		},
	}
}

var orderedMigrationToken = newOrderedMigrationToken()

func newOrderedMigrationToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Bird",
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Bird},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
		},
	}
}
