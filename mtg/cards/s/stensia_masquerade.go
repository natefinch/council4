package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// StensiaMasquerade is the card definition for Stensia Masquerade.
//
// Type: Enchantment
// Cost: {2}{R}
//
// Oracle text:
//
//	Attacking creatures you control have first strike.
//	Whenever a Vampire you control deals combat damage to a player, put a +1/+1 counter on it.
//	Madness {2}{R} (If you discard this card, discard it into exile. When you do, cast it for its madness cost or put it into your graveyard.)
var StensiaMasquerade = newStensiaMasquerade

func newStensiaMasquerade() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Stensia Masquerade",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Attacking creatures you control have first strike.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}, CombatState: game.CombatStateAttacking}),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.FirstStrike},
						},
					},
				},
				game.StaticAbility{
					Text: "Madness {2}{R}",
					KeywordAbilities: []game.KeywordAbility{
						game.MadnessKeyword{
							Cost: cost.Mana{
								cost.O(2),
								cost.R,
							},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever a Vampire you control deals combat damage to a player, put a +1/+1 counter on it.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                 game.EventDamageDealt,
							Controller:            game.TriggerControllerYou,
							Subject:               game.TriggerSubjectDamageSource,
							StackObjectKind:       game.StackSpell,
							DamageRecipient:       game.DamageRecipientPlayer,
							DamageSourceSelection: game.Selection{SubtypesAny: []types.Sub{types.Vampire}},
							RequireCombatDamage:   true,
							CounterKind:           counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Object:      game.EventPermanentReference(),
											CounterKind: counter.PlusOnePlusOne,
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
			Attacking creatures you control have first strike.
			Whenever a Vampire you control deals combat damage to a player, put a +1/+1 counter on it.
			Madness {2}{R} (If you discard this card, discard it into exile. When you do, cast it for its madness cost or put it into your graveyard.)
		`,
		},
	}
}
