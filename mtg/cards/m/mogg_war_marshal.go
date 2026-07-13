package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MoggWarMarshal is the card definition for Mogg War Marshal.
//
// Type: Creature — Goblin Warrior
// Cost: {1}{R}
//
// Oracle text:
//
//	Echo {1}{R} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	When this creature enters or dies, create a 1/1 red Goblin creature token.
var MoggWarMarshal = newMoggWarMarshal

func newMoggWarMarshal() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Mogg War Marshal",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Goblin, types.Warrior},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.EchoTriggeredAbility(cost.Mana{cost.O(1), cost.R}),
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
									Amount: game.Fixed(1),
									Source: game.TokenDef(moggWarMarshalToken),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Echo {1}{R} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
			When this creature enters or dies, create a 1/1 red Goblin creature token.
		`,
		},
	}
}

var moggWarMarshalToken = newMoggWarMarshalToken()

func newMoggWarMarshalToken() *game.CardDef {
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
