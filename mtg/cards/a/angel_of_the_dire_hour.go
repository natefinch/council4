package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AngelOfTheDireHour is the card definition for Angel of the Dire Hour.
//
// Type: Creature — Angel
// Cost: {5}{W}{W}
//
// Oracle text:
//
//	Flash
//	Flying
//	When this creature enters, if you cast it from your hand, exile all attacking creatures.
var AngelOfTheDireHour = newAngelOfTheDireHour

func newAngelOfTheDireHour() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Angel of the Dire Hour",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.W,
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Angel},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
				game.FlyingStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
						InterveningIf: "if you cast it from your hand",
						InterveningIfEventPermanentWasCastFromControllerHand: true,
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Exile{
									Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, CombatState: game.CombatStateAttacking}),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flash
			Flying
			When this creature enters, if you cast it from your hand, exile all attacking creatures.
		`,
		},
	}
}
