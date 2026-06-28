package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ResoundingScream is the card definition for Resounding Scream.
//
// Type: Sorcery
// Cost: {2}{B}
//
// Oracle text:
//
//	Target player discards a card at random.
//	Cycling {5}{U}{B}{R} ({5}{U}{B}{R}, Discard this card: Draw a card.)
//	When you cycle this card, target player discards two cards at random.
var ResoundingScream = newResoundingScream()

func newResoundingScream() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Resounding Scream",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(5), cost.U, cost.B, cost.R}),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventCycled,
							Source: game.TriggerSourceSelf,
							Player: game.TriggerPlayerYou,
						},
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target player",
								Allow:      game.TargetAllowPlayer,
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Discard{
									Amount:   game.Fixed(2),
									Player:   game.TargetPlayerReference(0),
									AtRandom: true,
								},
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
						Constraint: "Target player",
						Allow:      game.TargetAllowPlayer,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Discard{
							Amount:   game.Fixed(1),
							Player:   game.TargetPlayerReference(0),
							AtRandom: true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Target player discards a card at random.
			Cycling {5}{U}{B}{R} ({5}{U}{B}{R}, Discard this card: Draw a card.)
			When you cycle this card, target player discards two cards at random.
		`,
		},
	}
}
