package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HERBIELovableRobot is the card definition for H.E.R.B.I.E., Lovable Robot.
//
// Type: Legendary Artifact Creature — Robot Scout
// Cost: {2}
//
// Oracle text:
//
//	Flying
//	At the beginning of combat on your turn, if you've cast a noncreature spell this turn, surveil 1.
//	{T}: Add {C}.
//	{1}, {T}: Add one mana of any color.
var HERBIELovableRobot = newHERBIELovableRobot

func newHERBIELovableRobot() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "H.E.R.B.I.E., Lovable Robot",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact, types.Creature},
			Subtypes:   []types.Sub{types.Robot, types.Scout},
			Power:      opt.Val(game.PT{Value: 1}),
			Toughness:  opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.C),
				game.ManaAbility{
					ManaCost:        opt.Val(cost.Mana{cost.O(1)}),
					AdditionalCosts: cost.Tap,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Choose{
									Choice: game.ResolutionChoice{
										Kind:   game.ResolutionChoiceMana,
										Prompt: "Choose a color",
										Colors: []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G},
									},
									PublishChoice: game.ChoiceKey("oracle-mana-color"),
								},
							},
							{
								Primitive: game.AddMana{
									Amount:     game.Fixed(1),
									ChoiceFrom: game.ChoiceKey("oracle-mana-color"),
								},
							},
						},
					}.Ability(),
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepBeginningOfCombat,
						},
						InterveningIf: "if you've cast a noncreature spell this turn",
						InterveningCondition: opt.Val(game.Condition{
							EventHistory: opt.Val(game.EventHistoryCondition{Pattern: game.TriggerPattern{
								Event:         game.EventSpellCast,
								Controller:    game.TriggerControllerYou,
								CardSelection: game.Selection{ExcludedTypes: []types.Card{types.Creature}},
							}, Window: game.EventHistoryCurrentTurn}),
						}),
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Surveil{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flying
			At the beginning of combat on your turn, if you've cast a noncreature spell this turn, surveil 1.
			{T}: Add {C}.
			{1}, {T}: Add one mana of any color.
		`,
		},
	}
}
