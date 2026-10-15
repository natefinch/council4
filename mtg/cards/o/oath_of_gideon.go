package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// OathOfGideon is the card definition for Oath of Gideon.
//
// Type: Legendary Enchantment
// Cost: {2}{W}
//
// Oracle text:
//
//	When Oath of Gideon enters, create two 1/1 white Kor Ally creature tokens.
//	Each planeswalker you control enters with an additional loyalty counter on it.
var OathOfGideon = newOathOfGideon

func newOathOfGideon() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Oath of Gideon",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Enchantment},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When Oath of Gideon enters, create two 1/1 white Kor Ally creature tokens.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentEnteredBattlefield,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(2),
											Source: game.TokenDef(oathOfGideonToken),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersGroupReplacement("Each planeswalker you control enters with an additional loyalty counter on it.", &game.Selection{RequiredTypes: []types.Card{types.Planeswalker}, Controller: game.ControllerYou}, game.CounterPlacement{Kind: counter.Loyalty, Amount: 1}),
			},
			OracleText: `
			When Oath of Gideon enters, create two 1/1 white Kor Ally creature tokens.
			Each planeswalker you control enters with an additional loyalty counter on it.
		`,
		},
	}
}

var oathOfGideonToken = newOathOfGideonToken()

func newOathOfGideonToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Kor Ally",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Kor, types.Ally},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
