package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// FlamewarBrashVeteran is the card definition for Flamewar, Brash Veteran // Flamewar, Streetwise Operative.
//
// Type: Legendary Artifact Creature — Robot // Legendary Artifact — Vehicle
// Face: Flamewar, Streetwise Operative — Legendary Artifact — Vehicle
//
// Oracle text:
//
//	More Than Meets the Eye {B}{R} (You may cast this card converted for {B}{R}.)
//	Sacrifice another artifact: Put a +1/+1 counter on Flamewar and convert it. Activate only as a sorcery.
//	{1}, Discard your hand: Put all exiled cards you own with intel counters on them into your hand.
var FlamewarBrashVeteran = newFlamewarBrashVeteran

func newFlamewarBrashVeteran() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Flamewar, Brash Veteran",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
				cost.R,
			}),
			Colors:     []color.Color{color.Black, color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact, types.Creature},
			Subtypes:   []types.Sub{types.Robot},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Sacrifice another artifact: Put a +1/+1 counter on Flamewar and convert it. Activate only as a sorcery.",
					AdditionalCosts: []cost.Additional{
						{
							Kind:               cost.AdditionalSacrifice,
							Text:               "Sacrifice another artifact",
							Amount:             1,
							MatchPermanentType: true,
							PermanentType:      types.Artifact,
							ExcludeSource:      true,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Timing:         game.SorceryOnly,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.SourcePermanentReference(),
									CounterKind: counter.PlusOnePlusOne,
								},
							},
							{
								Primitive: game.Transform{
									Object: game.SourcePermanentReference(),
								},
							},
						},
					}.Ability(),
				},
				game.ActivatedAbility{
					Text:     "{1}, Discard your hand: Put all exiled cards you own with intel counters on them into your hand.",
					ManaCost: opt.Val(cost.Mana{cost.O(1)}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:          cost.AdditionalDiscard,
							Text:          "Discard your hand",
							AmountDynamic: cost.AdditionalDynamicHandSize,
							Source:        zone.Hand,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ReturnExiledCardsWithCounter{
									Player:  game.ControllerReference(),
									Counter: counter.Intel,
								},
							},
						},
					}.Ability(),
				},
			},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label:    "More Than Meets the Eye",
					ManaCost: opt.Val(cost.Mana{cost.B, cost.R}),
					Mechanic: cost.AlternativeMechanicMoreThanMeetsTheEye,
				},
			},
			OracleText: `
			More Than Meets the Eye {B}{R} (You may cast this card converted for {B}{R}.)
			Sacrifice another artifact: Put a +1/+1 counter on Flamewar and convert it. Activate only as a sorcery.
			{1}, Discard your hand: Put all exiled cards you own with intel counters on them into your hand.
		`,
		},
		Layout: game.LayoutTransform,
		Back: opt.Val(game.CardFace{
			Name:       "Flamewar, Streetwise Operative",
			Colors:     []color.Color{color.Black, color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact},
			Subtypes:   []types.Sub{types.Vehicle},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.LivingMetalStaticBody,
				game.MenaceStaticBody,
				game.DeathtouchStaticBody,
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
								Primitive: game.ExileTopOfLibrary{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountEventDamage,
										Multiplier: 1,
									}),
									Player:   game.ControllerReference(),
									Counter:  opt.Val(counter.Intel),
									FaceDown: true,
								},
							},
							{
								Primitive: game.Transform{
									Object: game.SourcePermanentReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Living metal (During your turn, this Vehicle is also a creature.)
			Menace, deathtouch
			Whenever Flamewar deals combat damage to a player, exile that many cards from the top of your library face down. Put an intel counter on each of them. Convert Flamewar.
		`,
		}),
	}
}
