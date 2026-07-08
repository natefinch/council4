package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SphinxOfUthuun is the card definition for Sphinx of Uthuun.
//
// Type: Creature — Sphinx
// Cost: {5}{U}{U}
//
// Oracle text:
//
//	Flying
//	When this creature enters, reveal the top five cards of your library. An opponent separates those cards into two piles. Put one pile into your hand and the other into your graveyard.
var SphinxOfUthuun = newSphinxOfUthuun

func newSphinxOfUthuun() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Sphinx of Uthuun",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.U,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Sphinx},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
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
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.PileSplit{
									Player:            game.ControllerReference(),
									Amount:            game.Fixed(5),
									SeparatorOpponent: true,
									Kept:              zone.Hand,
									Other:             zone.Graveyard,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flying
			When this creature enters, reveal the top five cards of your library. An opponent separates those cards into two piles. Put one pile into your hand and the other into your graveyard.
		`,
		},
	}
}
