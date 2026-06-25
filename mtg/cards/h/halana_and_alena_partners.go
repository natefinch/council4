package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"

	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HalanaAndAlenaPartners is the card definition for Halana and Alena, Partners.
//
// Type: Legendary Creature — Human Ranger
// Cost: {2}{R}{G}
//
// Oracle text:
//
//	First strike (This creature deals combat damage before creatures without first strike.)
//	Reach (This creature can block creatures with flying.)
//	At the beginning of combat on your turn, put X +1/+1 counters on another target creature you control, where X is Halana and Alena's power. That creature gains haste until end of turn.
var HalanaAndAlenaPartners = func() *game.CardDef {
	card := &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green, color.Red),
		CardFace: game.CardFace{
			Name: "Halana and Alena, Partners",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
				cost.G,
			}),
			Colors:     []color.Color{color.Green, color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Ranger},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			OracleText: `
				First strike (This creature deals combat damage before creatures without first strike.)
				Reach (This creature can block creatures with flying.)
				At the beginning of combat on your turn, put X +1/+1 counters on another target creature you control, where X is Halana and Alena's power. That creature gains haste until end of turn.
			`,
		},
	}

	card.StaticAbilities = append(card.StaticAbilities,
		game.FirstStrikeStaticBody,
	)

	card.StaticAbilities = append(card.StaticAbilities,
		game.ReachStaticBody,
	)

	card.TriggeredAbilities = append(card.TriggeredAbilities,
		game.TriggeredAbility{
			Text: `
				At the beginning of combat on your turn, put X +1/+1 counters on another target creature you control, where X is Halana and Alena's power. That creature gains haste until end of turn.
			`,
			Trigger: game.TriggerCondition{
				Type: game.TriggerAt,
				Pattern: game.TriggerPattern{
					Event:      game.EventBeginningOfStep,
					Controller: game.TriggerControllerYou,
					Step:       game.StepBeginningOfCombat,
				},
			},
			Content: game.Mode{
				Targets: []game.TargetSpec{
					{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "another creature you control",
						Allow:      game.TargetAllowPermanent,
						Selection: opt.Val(game.Selection{
							RequiredTypesAny: []types.Card{
								types.Creature,
							},
							Controller:    game.ControllerYou,
							ExcludeSource: true,
						}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.AddCounter{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:   game.DynamicAmountTargetPower,
								Object: game.SourcePermanentReference(),
							}),
							Object:      game.TargetPermanentReference(0),
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					{
						Primitive: game.ApplyContinuous{
							Object: opt.Val(game.TargetPermanentReference(0)),
							ContinuousEffects: []game.ContinuousEffect{
								{
									Layer: game.LayerAbility,
									AddKeywords: []game.Keyword{
										game.Haste,
									},
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
				},
			}.Ability(),
		},
	)
	return card
}()
