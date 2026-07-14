package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BraidsArisenNightmare is the card definition for Braids, Arisen Nightmare.
//
// Type: Legendary Creature — Nightmare
// Cost: {1}{B}{B}
//
// Oracle text:
//
//	At the beginning of your end step, you may sacrifice an artifact, creature, enchantment, land, or planeswalker. If you do, each opponent may sacrifice a permanent of their choice that shares a card type with it. For each opponent who doesn't, that player loses 2 life and you draw a card.
var BraidsArisenNightmare = newBraidsArisenNightmare

func newBraidsArisenNightmare() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Braids, Arisen Nightmare",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
				cost.B,
			}),
			Colors:     []color.Color{color.Black},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Nightmare},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepEnd,
						},
					},
					Content: game.Mode{
						Text: "At the beginning of your end step, you may sacrifice an artifact, creature, enchantment, land, or planeswalker. If you do, each opponent may sacrifice a permanent of their choice that shares a card type with it. For each opponent who doesn't, that player loses 2 life and you draw a card.",
						Sequence: []game.Instruction{
							{
								Primitive: game.SacrificePermanents{
									Amount:               game.Fixed(1),
									Player:               game.ControllerReference(),
									Selection:            game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature, types.Enchantment, types.Land, types.Planeswalker}},
									PublishLinked:        game.LinkedKey("braids-sacrificed-permanent"),
									PublishObjectBinding: true,
								},
								Optional:      true,
								PublishResult: game.ResultKey("braids-sacrificed"),
							},
							{
								Primitive: game.PunisherEachLoseLife{
									PlayerGroup:        game.OpponentsReference(),
									Amount:             game.Fixed(2),
									AllowSacrifice:     true,
									SacrificeSelection: game.Selection{SharesCardTypeFromLinked: game.LinkedKey("braids-sacrificed-permanent")},
									ControllerDrawEach: true,
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "braids-sacrificed",
									Succeeded: game.TriTrue,
								}),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			At the beginning of your end step, you may sacrifice an artifact, creature, enchantment, land, or planeswalker. If you do, each opponent may sacrifice a permanent of their choice that shares a card type with it. For each opponent who doesn't, that player loses 2 life and you draw a card.
		`,
		},
	}
}
