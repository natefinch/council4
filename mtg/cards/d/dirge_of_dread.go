package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DirgeOfDread is the card definition for Dirge of Dread.
//
// Type: Sorcery
// Cost: {2}{B}
//
// Oracle text:
//
//	All creatures gain fear until end of turn. (They can't be blocked except by artifact creatures and/or black creatures.)
//	Cycling {1}{B} ({1}{B}, Discard this card: Draw a card.)
//	When you cycle this card, you may have target creature gain fear until end of turn.
var DirgeOfDread = newDirgeOfDread

func newDirgeOfDread() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Dirge of Dread",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(1), cost.B}),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When you cycle this card, you may have target creature gain fear until end of turn.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventCycled,
							Source:          game.TriggerSourceSelf,
							Player:          game.TriggerPlayerYou,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
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
										Primitive: game.ApplyContinuous{
											Object: opt.Val(game.TargetPermanentReference(0)),
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Duration:    game.DurationPermanent,
													Layer:       game.LayerAbility,
													AddKeywords: []game.Keyword{game.Fear},
												},
											},
											Duration: game.DurationUntilEndOfTurn,
										},
										Optional: true,
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.ApplyContinuous{
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Duration:    game.DurationPermanent,
									Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
									Layer:       game.LayerAbility,
									AddKeywords: []game.Keyword{game.Fear},
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			All creatures gain fear until end of turn. (They can't be blocked except by artifact creatures and/or black creatures.)
			Cycling {1}{B} ({1}{B}, Discard this card: Draw a card.)
			When you cycle this card, you may have target creature gain fear until end of turn.
		`,
		},
	}
}
