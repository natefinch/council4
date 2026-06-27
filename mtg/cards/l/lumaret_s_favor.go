package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LumaretSFavor is the card definition for Lumaret's Favor.
//
// Type: Instant
// Cost: {1}{G}
//
// Oracle text:
//
//	Infusion — When you cast this spell, copy it if you gained life this turn. You may choose new targets for the copy.
//	Target creature gets +2/+4 until end of turn.
var LumaretSFavor = newLumaretSFavor()

func newLumaretSFavor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Lumaret's Favor",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Instant},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:       game.EventSpellCast,
							Source:      game.TriggerSourceSelf,
							Controller:  game.TriggerControllerYou,
							SelfWasCast: true,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CopyStackObject{
									Object:              game.EventStackObjectReference(),
									MayChooseNewTargets: true,
								},
								Condition: opt.Val(game.EffectCondition{
									Condition: opt.Val(game.Condition{
										EventHistory: opt.Val(game.EventHistoryCondition{Pattern: game.TriggerPattern{
											Event:  game.EventLifeGained,
											Player: game.TriggerPlayerYou,
										}, Window: game.EventHistoryCurrentTurn}),
									}),
								}),
							},
						},
					}.Ability(),
				},
			},
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
						Primitive: game.ModifyPT{
							Object:         game.TargetPermanentReference(0),
							PowerDelta:     game.Fixed(2),
							ToughnessDelta: game.Fixed(4),
							Duration:       game.DurationUntilEndOfTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Infusion — When you cast this spell, copy it if you gained life this turn. You may choose new targets for the copy.
			Target creature gets +2/+4 until end of turn.
		`,
		},
	}
}
