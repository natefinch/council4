package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MentorSGuidance is the card definition for Mentor's Guidance.
//
// Type: Sorcery
// Cost: {2}{U}
//
// Oracle text:
//
//	When you cast this spell, copy it if you control a planeswalker, Cleric, Druid, Shaman, Warlock, or Wizard.
//	Scry 1, then draw a card.
var MentorSGuidance = newMentorSGuidance()

func newMentorSGuidance() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Mentor's Guidance",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
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
									Object: game.EventStackObjectReference(),
								},
								Condition: opt.Val(game.EffectCondition{
									Condition: opt.Val(game.Condition{
										ControlsMatching: opt.Val(game.SelectionCount{
											Selection: game.Selection{RequiredTypes: []types.Card{types.Planeswalker}},
										}),
									}),
								}),
							},
						},
					}.Ability(),
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Scry{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
					},
					{
						Primitive: game.Draw{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			When you cast this spell, copy it if you control a planeswalker, Cleric, Druid, Shaman, Warlock, or Wizard.
			Scry 1, then draw a card.
		`,
		},
	}
}
