package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HarmoniousArchon is the card definition for Harmonious Archon.
//
// Type: Creature — Archon
// Cost: {4}{W}{W}
//
// Oracle text:
//
//	Flying
//	Non-Archon creatures have base power and toughness 3/3.
//	When this creature enters, create two 1/1 white Human creature tokens.
var HarmoniousArchon = newHarmoniousArchon()

func newHarmoniousArchon() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Harmonious Archon",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.W,
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Archon},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:        game.LayerPowerToughnessSet,
							Group:        game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Archon")}}),
							SetPower:     opt.Val(game.PT{Value: 3}),
							SetToughness: opt.Val(game.PT{Value: 3}),
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(2),
									Source: game.TokenDef(harmoniousArchonToken),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flying
			Non-Archon creatures have base power and toughness 3/3.
			When this creature enters, create two 1/1 white Human creature tokens.
		`,
		},
	}
}

var harmoniousArchonToken = newHarmoniousArchonToken()

func newHarmoniousArchonToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Human",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
