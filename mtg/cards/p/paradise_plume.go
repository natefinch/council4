package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ParadisePlume is the card definition for Paradise Plume.
//
// Type: Artifact
// Cost: {4}
//
// Oracle text:
//
//	As this artifact enters, choose a color.
//	Whenever a player casts a spell of the chosen color, you may gain 1 life.
//	{T}: Add one mana of the chosen color.
var ParadisePlume = newParadisePlume

func newParadisePlume() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Paradise Plume",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types: []types.Card{types.Artifact},
			ManaAbilities: []game.ManaAbility{
				game.TapChosenColorManaAbility("{T}: Add one mana of the chosen color."),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever a player casts a spell of the chosen color, you may gain 1 life.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventSpellCast,
							CardSelection:   game.Selection{ColorChoice: game.ColorChoiceSourceEntry},
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Optional: true,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.GainLife{
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
				game.EntryColorChoiceReplacement("As this artifact enters, choose a color."),
			},
			OracleText: `
			As this artifact enters, choose a color.
			Whenever a player casts a spell of the chosen color, you may gain 1 life.
			{T}: Add one mana of the chosen color.
		`,
		},
	}
}
