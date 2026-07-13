package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DerangedHermit is the card definition for Deranged Hermit.
//
// Type: Creature — Elf
// Cost: {3}{G}{G}
//
// Oracle text:
//
//	Echo {3}{G}{G} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	When this creature enters, create four 1/1 green Squirrel creature tokens.
//	Squirrel creatures get +1/+1.
var DerangedHermit = newDerangedHermit

func newDerangedHermit() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Deranged Hermit",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elf},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							Group:          game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Squirrel")}}),
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.EchoTriggeredAbility(cost.Mana{cost.O(3), cost.G, cost.G}),
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
									Amount: game.Fixed(4),
									Source: game.TokenDef(derangedHermitToken),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Echo {3}{G}{G} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
			When this creature enters, create four 1/1 green Squirrel creature tokens.
			Squirrel creatures get +1/+1.
		`,
		},
	}
}

var derangedHermitToken = newDerangedHermitToken()

func newDerangedHermitToken() *game.CardDef {
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
