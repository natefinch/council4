package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Aberrant is the card definition for Aberrant.
//
// Type: Creature — Tyranid Mutant
// Cost: {X}{1}{G}
//
// Oracle text:
//
//	Ravenous (This creature enters with X +1/+1 counters on it. If X is 5 or more, draw a card when it enters.)
//	Trample
//	Heavy Power Hammer — Whenever this creature deals combat damage to a player, destroy target artifact or enchantment that player controls.
var Aberrant = newAberrant

func newAberrant() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Aberrant",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.O(1),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Tyranid, types.Mutant},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 0}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.RavenousDrawTriggeredAbility(),
				game.TriggeredAbility{
					Text: "Heavy Power Hammer — Whenever this creature deals combat damage to a player, destroy target artifact or enchantment that player controls.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:               game.EventDamageDealt,
							Source:              game.TriggerSourceSelf,
							Subject:             game.TriggerSubjectDamageSource,
							StackObjectKind:     game.StackSpell,
							DamageRecipient:     game.DamageRecipientPlayer,
							RequireCombatDamage: true,
							CounterKind:         counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target artifact or enchantment that player controls",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Enchantment}, ControlledByEventPlayer: true}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Destroy{
											Object: game.TargetPermanentReference(0),
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
				game.RavenousEntersWithCountersReplacement(),
			},
			OracleText: `
			Ravenous (This creature enters with X +1/+1 counters on it. If X is 5 or more, draw a card when it enters.)
			Trample
			Heavy Power Hammer — Whenever this creature deals combat damage to a player, destroy target artifact or enchantment that player controls.
		`,
		},
	}
}
