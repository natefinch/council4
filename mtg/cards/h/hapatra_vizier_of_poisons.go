package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HapatraVizierOfPoisons is the card definition for Hapatra, Vizier of Poisons.
//
// Type: Legendary Creature — Human Cleric
// Cost: {B}{G}
//
// Oracle text:
//
//	Whenever Hapatra deals combat damage to a player, you may put a -1/-1 counter on target creature.
//	Whenever you put one or more -1/-1 counters on a creature, create a 1/1 green Snake creature token with deathtouch.
var HapatraVizierOfPoisons = newHapatraVizierOfPoisons

func newHapatraVizierOfPoisons() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Hapatra, Vizier of Poisons",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
				cost.G,
			}),
			Colors:     []color.Color{color.Black, color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Cleric},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever Hapatra deals combat damage to a player, you may put a -1/-1 counter on target creature.",
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
					Optional: true,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target creature",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Object:      game.TargetPermanentReference(0),
											CounterKind: counter.MinusOneMinusOne,
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.TriggeredAbility{
					Text: "Whenever you put one or more -1/-1 counters on a creature, create a 1/1 green Snake creature token with deathtouch.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventCountersAdded,
							CauseController:  game.TriggerControllerYou,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
							StackObjectKind:  game.StackSpell,
							OneOrMore:        true,
							MatchCounterKind: true,
							CounterKind:      counter.MinusOneMinusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(hapatraVizierOfPoisonsToken),
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
			Whenever Hapatra deals combat damage to a player, you may put a -1/-1 counter on target creature.
			Whenever you put one or more -1/-1 counters on a creature, create a 1/1 green Snake creature token with deathtouch.
		`,
		},
	}
}

var hapatraVizierOfPoisonsToken = newHapatraVizierOfPoisonsToken()

func newHapatraVizierOfPoisonsToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Snake",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Snake},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.DeathtouchStaticBody,
			},
		},
	}
}
