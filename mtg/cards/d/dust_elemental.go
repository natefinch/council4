package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DustElemental is the card definition for Dust Elemental.
//
// Type: Creature — Elemental
// Cost: {2}{W}{W}
//
// Oracle text:
//
//	Flash (You may cast this spell any time you could cast an instant.)
//	Flying; fear (This creature can't be blocked except by artifact creatures and/or black creatures.)
//	When this creature enters, return three creatures you control to their owner's hand.
var DustElemental = newDustElemental()

func newDustElemental() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Dust Elemental",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elemental},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
				game.FlyingStaticBody,
				game.FearStaticBody,
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
								Primitive: game.Bounce{
									ControlledChoice: true,
									Amount:           game.Fixed(3),
									Group:            game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flash (You may cast this spell any time you could cast an instant.)
			Flying; fear (This creature can't be blocked except by artifact creatures and/or black creatures.)
			When this creature enters, return three creatures you control to their owner's hand.
		`,
		},
	}
}
