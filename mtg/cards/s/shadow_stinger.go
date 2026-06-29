package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ShadowStinger is the card definition for Shadow Stinger.
//
// Type: Creature — Vampire Rogue
// Cost: {2}{B}
//
// Oracle text:
//
//	Tap another untapped Rogue you control: This creature gains deathtouch until end of turn.
//	Whenever this creature deals combat damage to a player, that player mills three cards. (They put the top three cards of their library into their graveyard.)
var ShadowStinger = newShadowStinger()

func newShadowStinger() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Shadow Stinger",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Vampire, types.Rogue},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 4}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Tap another untapped Rogue you control: This creature gains deathtouch until end of turn.",
					AdditionalCosts: []cost.Additional{
						{
							Kind:          cost.AdditionalTapPermanents,
							Text:          "Tap another untapped Rogue you control",
							Amount:        1,
							ExcludeSource: true,
							SubtypesAny:   cost.SubtypeSet{types.Rogue},
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									Object: opt.Val(game.SourceCardPermanentReference()),
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											AddKeywords: []game.Keyword{
												game.Deathtouch,
											},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
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
								Primitive: game.Mill{
									Amount: game.Fixed(3),
									Player: game.EventPlayerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Tap another untapped Rogue you control: This creature gains deathtouch until end of turn.
			Whenever this creature deals combat damage to a player, that player mills three cards. (They put the top three cards of their library into their graveyard.)
		`,
		},
	}
}
