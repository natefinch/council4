package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ParadoxEngine is the card definition for Paradox Engine.
//
// Type: Legendary Artifact
// Cost: {5}
//
// Oracle text:
//
//	Whenever you cast a spell, untap all nonland permanents you control.
var ParadoxEngine = newParadoxEngine

func newParadoxEngine() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Paradox Engine",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
			}),
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever you cast a spell, untap all nonland permanents you control.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventSpellCast,
							Controller:      game.TriggerControllerYou,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Untap{
											Group: game.BattlefieldGroup(game.Selection{ExcludedTypes: []types.Card{types.Land}, Controller: game.ControllerYou}),
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
			Whenever you cast a spell, untap all nonland permanents you control.
		`,
		},
	}
}
