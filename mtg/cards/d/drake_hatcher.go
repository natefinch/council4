package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// DrakeHatcher is the card definition for Drake Hatcher.
//
// Type: Creature — Human Wizard
// Cost: {1}{U}
//
// Oracle text:
//
//	Vigilance, prowess (Whenever you cast a noncreature spell, this creature gets +1/+1 until end of turn.)
//	Whenever this creature deals combat damage to a player, put that many incubation counters on it.
//	Remove three incubation counters from this creature: Create a 2/2 blue Drake creature token with flying.
var DrakeHatcher = newDrakeHatcher()

func newDrakeHatcher() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Drake Hatcher",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Wizard},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
				game.ProwessStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Remove three incubation counters from this creature: Create a 2/2 blue Drake creature token with flying.",
					AdditionalCosts: []cost.Additional{
						{
							Kind:        cost.AdditionalRemoveCounter,
							Text:        "Remove three incubation counters from this creature",
							Amount:      3,
							CounterKind: counter.Incubation,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(drakeHatcherToken),
								},
							},
						},
					}.Ability(),
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:               game.EventDamageDealt,
							Source:              game.TriggerSourceSelf,
							Subject:             game.TriggerSubjectDamageSource,
							RequireCombatDamage: true,
							DamageRecipient:     game.DamageRecipientPlayer,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountEventDamage,
										Multiplier: 1,
									}),
									Object:      game.EventPermanentReference(),
									CounterKind: counter.Incubation,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Vigilance, prowess (Whenever you cast a noncreature spell, this creature gets +1/+1 until end of turn.)
			Whenever this creature deals combat damage to a player, put that many incubation counters on it.
			Remove three incubation counters from this creature: Create a 2/2 blue Drake creature token with flying.
		`,
		},
	}
}

var drakeHatcherToken = newDrakeHatcherToken()

func newDrakeHatcherToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Drake",
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Drake},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
		},
	}
}
