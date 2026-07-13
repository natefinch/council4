package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TroublemakerOuphe is the card definition for Troublemaker Ouphe.
//
// Type: Creature — Ouphe
// Cost: {1}{G}
//
// Oracle text:
//
//	Bargain (You may sacrifice an artifact, enchantment, or token as you cast this spell.)
//	When this creature enters, if it was bargained, exile target artifact or enchantment an opponent controls.
var TroublemakerOuphe = newTroublemakerOuphe

func newTroublemakerOuphe() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Troublemaker Ouphe",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Ouphe},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.BargainStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
						InterveningIf:                           "if it was bargained",
						InterveningIfEventPermanentWasBargained: true,
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target artifact or enchantment an opponent controls",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Enchantment}, Controller: game.ControllerOpponent}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Exile{
									Object: game.TargetPermanentReference(0),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Bargain (You may sacrifice an artifact, enchantment, or token as you cast this spell.)
			When this creature enters, if it was bargained, exile target artifact or enchantment an opponent controls.
		`,
		},
	}
}
