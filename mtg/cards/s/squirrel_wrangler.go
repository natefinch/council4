package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SquirrelWrangler is the card definition for Squirrel Wrangler.
//
// Type: Creature — Human Druid
// Cost: {2}{G}{G}
//
// Oracle text:
//
//	{1}{G}, Sacrifice a land: Create two 1/1 green Squirrel creature tokens.
//	{1}{G}, Sacrifice a land: Squirrel creatures get +1/+1 until end of turn.
var SquirrelWrangler = newSquirrelWrangler

func newSquirrelWrangler() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Squirrel Wrangler",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Druid},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{1}{G}, Sacrifice a land: Create two 1/1 green Squirrel creature tokens.",
					ManaCost: opt.Val(cost.Mana{cost.O(1), cost.G}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:               cost.AdditionalSacrifice,
							Text:               "Sacrifice a land",
							Amount:             1,
							MatchPermanentType: true,
							PermanentType:      types.Land,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(2),
									Source: game.TokenDef(squirrelWranglerToken),
								},
							},
						},
					}.Ability(),
				},
				game.ActivatedAbility{
					Text:     "{1}{G}, Sacrifice a land: Squirrel creatures get +1/+1 until end of turn.",
					ManaCost: opt.Val(cost.Mana{cost.O(1), cost.G}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:               cost.AdditionalSacrifice,
							Text:               "Sacrifice a land",
							Amount:             1,
							MatchPermanentType: true,
							PermanentType:      types.Land,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer:          game.LayerPowerToughnessModify,
											Group:          game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Squirrel")}}),
											PowerDelta:     1,
											ToughnessDelta: 1,
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{1}{G}, Sacrifice a land: Create two 1/1 green Squirrel creature tokens.
			{1}{G}, Sacrifice a land: Squirrel creatures get +1/+1 until end of turn.
		`,
		},
	}
}

var squirrelWranglerToken = newSquirrelWranglerToken()

func newSquirrelWranglerToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Squirrel",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Squirrel},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
