package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TraxosScourgeEternal is the card definition for Traxos, Scourge Eternal.
//
// Type: Legendary Artifact Creature — Dragon Construct
// Cost: {4}
//
// Oracle text:
//
//	Trample
//	Traxos doesn't untap during your untap step.
//	Whenever you cast an artifact or creature spell, untap Traxos.
var TraxosScourgeEternal = newTraxosScourgeEternal

func newTraxosScourgeEternal() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Traxos, Scourge Eternal",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact, types.Creature},
			Subtypes:   []types.Sub{types.Dragon, types.Construct},
			Power:      opt.Val(game.PT{Value: 5}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
				game.StaticAbility{
					Text: "Traxos doesn't untap during your untap step.",
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
					Text: "Whenever you cast an artifact or creature spell, untap Traxos.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventSpellCast,
							Controller:      game.TriggerControllerYou,
							CardSelection:   game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature}},
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
			Trample
			Traxos doesn't untap during your untap step.
			Whenever you cast an artifact or creature spell, untap Traxos.
		`,
		},
	}
}
