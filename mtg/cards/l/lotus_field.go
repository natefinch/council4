package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// LotusField is the card definition for Lotus Field.
//
// Type: Land
//
// Oracle text:
//
//	Hexproof
//	This land enters tapped.
//	When this land enters, sacrifice two lands.
//	{T}: Add three mana of any one color.
var LotusField = newLotusField

func newLotusField() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "Lotus Field",
			Types: []types.Card{types.Land},
			StaticAbilities: []game.StaticAbility{
				game.HexproofStaticBody,
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceCountAbility("{T}: Add three mana of any one color.", 3, mana.W, mana.U, mana.B, mana.R, mana.G),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this land enters, sacrifice two lands.",
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
										Primitive: game.SacrificePermanents{
											Player:    game.ControllerReference(),
											Amount:    game.Fixed(2),
											Selection: game.Selection{RequiredTypes: []types.Card{types.Land}},
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
			Hexproof
			This land enters tapped.
			When this land enters, sacrifice two lands.
			{T}: Add three mana of any one color.
		`,
		},
	}
}
