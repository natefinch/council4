package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// UndergroundMortuary is the card definition for Underground Mortuary.
//
// Type: Land — Swamp Forest
//
// Oracle text:
//
//	({T}: Add {B} or {G}.)
//	This land enters tapped.
//	When this land enters, surveil 1. (Look at the top card of your library. You may put it into your graveyard.)
var UndergroundMortuary = newUndergroundMortuary

func newUndergroundMortuary() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name:     "Underground Mortuary",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Swamp, types.Forest},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.B, mana.G),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this land enters, surveil 1. (Look at the top card of your library. You may put it into your graveyard.)",
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
										Primitive: game.Surveil{
											Amount: game.Fixed(1),
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
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			({T}: Add {B} or {G}.)
			This land enters tapped.
			When this land enters, surveil 1. (Look at the top card of your library. You may put it into your graveyard.)
		`,
		},
	}
}
