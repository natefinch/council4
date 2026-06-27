package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DenizenOfTheDeep is the card definition for Denizen of the Deep.
//
// Type: Creature — Serpent
// Cost: {6}{U}{U}
//
// Oracle text:
//
//	When this creature enters, return each other creature you control to its owner's hand.
var DenizenOfTheDeep = newDenizenOfTheDeep()

func newDenizenOfTheDeep() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Denizen of the Deep",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
				cost.U,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Serpent},
			Power:     opt.Val(game.PT{Value: 11}),
			Toughness: opt.Val(game.PT{Value: 11}),
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
									Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou, ExcludeSource: true}),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this creature enters, return each other creature you control to its owner's hand.
		`,
		},
	}
}
