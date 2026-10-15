package z

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ZenithChronicler is the card definition for Zenith Chronicler.
//
// Type: Artifact Creature — Phyrexian Construct
// Cost: {2}
//
// Oracle text:
//
//	Whenever a player casts their first multicolored spell each turn, each other player draws a card.
var ZenithChronicler = newZenithChronicler

func newZenithChronicler() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Zenith Chronicler",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian, types.Construct},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 1}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever a player casts their first multicolored spell each turn, each other player draws a card.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                      game.EventSpellCast,
							CardSelection:              game.Selection{Multicolored: true},
							StackObjectKind:            game.StackSpell,
							PlayerEventOrdinalThisTurn: 1,
							CounterKind:                counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Draw{
											Amount:      game.Fixed(1),
											PlayerGroup: game.OpponentsReference(),
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
			Whenever a player casts their first multicolored spell each turn, each other player draws a card.
		`,
		},
	}
}
