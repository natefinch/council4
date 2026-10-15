package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SphinxSovereign is the card definition for Sphinx Sovereign.
//
// Type: Artifact Creature — Sphinx
// Cost: {4}{W}{U}{U}{B}
//
// Oracle text:
//
//	Flying
//	At the beginning of your end step, you gain 3 life if this creature is untapped. Otherwise, each opponent loses 3 life.
var SphinxSovereign = newSphinxSovereign

func newSphinxSovereign() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Sphinx Sovereign",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.W,
				cost.U,
				cost.U,
				cost.B,
			}),
			Colors:    []color.Color{color.Black, color.Blue, color.White},
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Sphinx},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "At the beginning of your end step, you gain 3 life if this creature is untapped. Otherwise, each opponent loses 3 life.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:           game.EventBeginningOfStep,
							Controller:      game.TriggerControllerYou,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
							Step:            game.StepEnd,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.GainLife{
											Amount: game.Fixed(3),
											Player: game.ControllerReference(),
										},
										Condition: opt.Val(game.EffectCondition{
											Condition: opt.Val(game.Condition{
												Text:              "if this creature is untapped",
												Object:            opt.Val(game.SourcePermanentReference()),
												ObjectMatches:     opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, Tapped: game.TriFalse}),
												SourceCounterKind: counter.PlusOnePlusOne,
											}),
										}),
										PublishCondition: game.ConditionKey("condition-0"),
									},
									game.Instruction{
										Primitive: game.LoseLife{
											Amount:      game.Fixed(3),
											PlayerGroup: game.OpponentsReference(),
										},
										ConditionGate:       game.ConditionKey("condition-0"),
										ConditionGateNegate: true,
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
			Flying
			At the beginning of your end step, you gain 3 life if this creature is untapped. Otherwise, each opponent loses 3 life.
		`,
		},
	}
}
