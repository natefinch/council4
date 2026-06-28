package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DreadCacodemon is the card definition for Dread Cacodemon.
//
// Type: Creature — Demon
// Cost: {7}{B}{B}{B}
//
// Oracle text:
//
//	When this creature enters, if you cast it from your hand, destroy all creatures your opponents control, then tap all other creatures you control.
var DreadCacodemon = newDreadCacodemon()

func newDreadCacodemon() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Dread Cacodemon",
			ManaCost: opt.Val(cost.Mana{
				cost.O(7),
				cost.B,
				cost.B,
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Demon},
			Power:     opt.Val(game.PT{Value: 8}),
			Toughness: opt.Val(game.PT{Value: 8}),
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
								Primitive: game.Destroy{
									Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerOpponent}),
								},
							},
							{
								Primitive: game.Tap{
									Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou, ExcludeSource: true}),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this creature enters, if you cast it from your hand, destroy all creatures your opponents control, then tap all other creatures you control.
		`,
		},
	}
}
