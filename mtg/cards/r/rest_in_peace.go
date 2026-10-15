package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RestInPeace is the card definition for Rest in Peace.
//
// Type: Enchantment
// Cost: {1}{W}
//
// Oracle text:
//
//	When this enchantment enters, exile all graveyards.
//	If a card or token would be put into a graveyard from anywhere, exile it instead.
var RestInPeace = newRestInPeace

func newRestInPeace() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Rest in Peace",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Enchantment},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this enchantment enters, exile all graveyards.",
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
										Primitive: game.MoveCard{
											PlayerGroup: game.AllPlayersReference(),
											FromZone:    zone.Graveyard,
											Destination: zone.Exile,
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
				game.GraveyardRedirectReplacement("If a card or token would be put into a graveyard from anywhere, exile it instead.", game.TriggerControllerAny, game.TriggerControllerAny, false),
			},
			OracleText: `
			When this enchantment enters, exile all graveyards.
			If a card or token would be put into a graveyard from anywhere, exile it instead.
		`,
		},
	}
}
