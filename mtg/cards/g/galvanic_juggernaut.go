package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GalvanicJuggernaut is the card definition for Galvanic Juggernaut.
//
// Type: Artifact Creature — Juggernaut
// Cost: {4}
//
// Oracle text:
//
//	This creature attacks each combat if able.
//	This creature doesn't untap during your untap step.
//	Whenever another creature dies, untap this creature.
var GalvanicJuggernaut = newGalvanicJuggernaut

func newGalvanicJuggernaut() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Galvanic Juggernaut",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Juggernaut},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.MustAttackStaticBody,
				game.StaticAbility{
					Text: "This creature doesn't untap during your untap step.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectDoesntUntap,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever another creature dies, untap this creature.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentDied,
							ExcludeSelf:      true,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
							StackObjectKind:  game.StackSpell,
							CounterKind:      counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Untap{
											Object: game.SourcePermanentReference(),
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
			This creature attacks each combat if able.
			This creature doesn't untap during your untap step.
			Whenever another creature dies, untap this creature.
		`,
		},
	}
}
