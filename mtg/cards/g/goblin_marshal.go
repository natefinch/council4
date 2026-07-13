package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GoblinMarshal is the card definition for Goblin Marshal.
//
// Type: Creature — Goblin Warrior
// Cost: {4}{R}{R}
//
// Oracle text:
//
//	Echo {4}{R}{R} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	When this creature enters or dies, create two 1/1 red Goblin creature tokens.
var GoblinMarshal = newGoblinMarshal

func newGoblinMarshal() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Goblin Marshal",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.R,
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Goblin, types.Warrior},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.EchoTriggeredAbility(cost.Mana{cost.O(4), cost.R, cost.R}),
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:      game.EventPermanentEnteredBattlefield,
							Source:     game.TriggerSourceSelf,
							UnionEvent: game.EventPermanentDied,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(2),
									Source: game.TokenDef(goblinMarshalToken),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Echo {4}{R}{R} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
			When this creature enters or dies, create two 1/1 red Goblin creature tokens.
		`,
		},
	}
}

var goblinMarshalToken = newGoblinMarshalToken()

func newGoblinMarshalToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Goblin",
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Goblin},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
