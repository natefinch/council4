package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RallyAtTheHornburg is the card definition for Rally at the Hornburg.
//
// Type: Sorcery
// Cost: {1}{R}
//
// Oracle text:
//
//	Create two 1/1 white Human Soldier creature tokens. Humans you control gain haste until end of turn.
var RallyAtTheHornburg = newRallyAtTheHornburg

func newRallyAtTheHornburg() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Rally at the Hornburg",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(2),
							Source: game.TokenDef(rallyAtTheHornburgToken),
						},
					},
					{
						Primitive: game.ApplyContinuous{
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Layer: game.LayerAbility,
									Group: game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Human")}, Controller: game.ControllerYou}),
									AddKeywords: []game.Keyword{
										game.Haste,
									},
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create two 1/1 white Human Soldier creature tokens. Humans you control gain haste until end of turn.
		`,
		},
	}
}

var rallyAtTheHornburgToken = newRallyAtTheHornburgToken()

func newRallyAtTheHornburgToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Human Soldier",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Soldier},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
