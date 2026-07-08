package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// NestOfScarabs is the card definition for Nest of Scarabs.
//
// Type: Enchantment
//
// Oracle text:
//
//	Whenever you put one or more -1/-1 counters on a creature, create that many 1/1 black Insect creature tokens.
var NestOfScarabs = newNestOfScarabs

func newNestOfScarabs() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Nest of Scarabs",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Enchantment},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventCountersAdded,
							CauseController:  game.TriggerControllerYou,
							OneOrMore:        true,
							MatchCounterKind: true,
							CounterKind:      counter.MinusOneMinusOne,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountEventCounterCount,
										Multiplier: 1,
									}),
									Source: game.TokenDef(nestOfScarabsToken),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever you put one or more -1/-1 counters on a creature, create that many 1/1 black Insect creature tokens.
		`,
		},
	}
}

var nestOfScarabsToken = newNestOfScarabsToken()

func newNestOfScarabsToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Insect",
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Insect},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
