package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FlurryOfWings is the card definition for Flurry of Wings.
//
// Type: Instant
// Cost: {G}{W}{U}
//
// Oracle text:
//
//	Create X 1/1 white Bird Soldier creature tokens with flying, where X is the number of attacking creatures.
var FlurryOfWings = newFlurryOfWings()

func newFlurryOfWings() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Green),
		CardFace: game.CardFace{
			Name: "Flurry of Wings",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
				cost.W,
				cost.U,
			}),
			Colors: []color.Color{color.Green, color.Blue, color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountCountSelector,
								Multiplier: 1,
								Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, CombatState: game.CombatStateAttacking}),
							}),
							Source: game.TokenDef(flurryOfWingsToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create X 1/1 white Bird Soldier creature tokens with flying, where X is the number of attacking creatures.
		`,
		},
	}
}

var flurryOfWingsToken = newFlurryOfWingsToken()

func newFlurryOfWingsToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Bird Soldier",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Bird, types.Soldier},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
		},
	}
}
