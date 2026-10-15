package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CrypticAnnelid is the card definition for Cryptic Annelid.
//
// Type: Creature — Worm Beast
// Cost: {3}{U}
//
// Oracle text:
//
//	When this creature enters, scry 1, then scry 2, then scry 3. (To scry X, look at the top X cards of your library, then put any number of them on the bottom and the rest on top in any order.)
var CrypticAnnelid = newCrypticAnnelid

func newCrypticAnnelid() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Cryptic Annelid",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Worm, types.Beast},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 4}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this creature enters, scry 1, then scry 2, then scry 3. (To scry X, look at the top X cards of your library, then put any number of them on the bottom and the rest on top in any order.)",
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
										Primitive: game.Scry{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
										},
									},
									game.Instruction{
										Primitive: game.Scry{
											Amount: game.Fixed(2),
											Player: game.ControllerReference(),
										},
									},
									game.Instruction{
										Primitive: game.Scry{
											Amount: game.Fixed(3),
											Player: game.ControllerReference(),
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
			OracleText: `
			When this creature enters, scry 1, then scry 2, then scry 3. (To scry X, look at the top X cards of your library, then put any number of them on the bottom and the rest on top in any order.)
		`,
		},
	}
}
