package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WakeningSunSAvatar is the card definition for Wakening Sun's Avatar.
//
// Type: Creature — Dinosaur Avatar
// Cost: {5}{W}{W}{W}
//
// Oracle text:
//
//	When this creature enters, if you cast it from your hand, destroy all non-Dinosaur creatures.
var WakeningSunSAvatar = newWakeningSunSAvatar

func newWakeningSunSAvatar() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Wakening Sun's Avatar",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.W,
				cost.W,
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dinosaur, types.Avatar},
			Power:     opt.Val(game.PT{Value: 7}),
			Toughness: opt.Val(game.PT{Value: 7}),
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
									Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, ExcludedSubtype: types.Sub("Dinosaur")}),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this creature enters, if you cast it from your hand, destroy all non-Dinosaur creatures.
		`,
		},
	}
}
