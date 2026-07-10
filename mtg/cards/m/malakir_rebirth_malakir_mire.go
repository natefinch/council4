package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MalakirRebirth is the card definition for Malakir Rebirth // Malakir Mire.
//
// Type: Instant // Land
// Face: Malakir Mire — Land
//
// Oracle text:
//
//	Choose target creature. You lose 2 life. Until end of turn, that creature gains "When this creature dies, return it to the battlefield tapped under its owner's control."
var MalakirRebirth = newMalakirRebirth

func newMalakirRebirth() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Malakir Rebirth",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
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
					{
						Primitive: game.LoseLife{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
					},
					{
						Primitive: game.ApplyContinuous{
							Object: opt.Val(game.TargetPermanentReference(0)),
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Layer: game.LayerAbility,
									AddAbilities: []game.Ability{
										new(game.TriggeredAbility{
											Trigger: game.TriggerCondition{
												Type: game.TriggerWhen,
												Pattern: game.TriggerPattern{
													Event:            game.EventPermanentDied,
													Source:           game.TriggerSourceSelf,
													SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
												},
											},
											Content: game.Mode{
												Sequence: []game.Instruction{
													{
														Primitive: game.PutOnBattlefield{
															Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceEvent}),
														},
													},
												},
											}.Ability(),
										}),
									},
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Choose target creature. You lose 2 life. Until end of turn, that creature gains "When this creature dies, return it to the battlefield tapped under its owner's control."
		`,
		},
		Layout: game.LayoutModalDFC,
		Back: opt.Val(game.CardFace{
			Name:  "Malakir Mire",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.B),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			This land enters tapped.
			{T}: Add {B}.
		`,
		}),
	}
}
