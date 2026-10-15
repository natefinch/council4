package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BiteOfTheBlackRose is the card definition for Bite of the Black Rose.
//
// Type: Sorcery
// Cost: {3}{B}
//
// Oracle text:
//
//	Will of the council — Starting with you, each player votes for sickness or psychosis. If sickness gets more votes, creatures your opponents control get -2/-2 until end of turn. If psychosis gets more votes or the vote is tied, each opponent discards two cards.
var BiteOfTheBlackRose = newBiteOfTheBlackRose

func newBiteOfTheBlackRose() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Bite of the Black Rose",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Vote{
							Options: []string{"sickness", "psychosis"},
						},
						PublishResult: game.ResultKey("vote-result"),
					},
					{
						Primitive: game.ApplyContinuous{
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Duration:       game.DurationPermanent,
									Group:          game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerOpponent}),
									Layer:          game.LayerPowerToughnessModify,
									PowerDelta:     -2,
									ToughnessDelta: -2,
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
						ResultGate: opt.Val(game.InstructionResultGate{
							Key:         "vote-result",
							AmountRange: opt.Val(game.IntRange{Min: 1, Max: 4}),
						}),
					},
					{
						Primitive: game.Discard{
							Amount:      game.Fixed(2),
							PlayerGroup: game.OpponentsReference(),
						},
						ResultGate: opt.Val(game.InstructionResultGate{
							Key:         "vote-result",
							AmountRange: opt.Val(game.IntRange{Min: -4, Max: 0}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Will of the council — Starting with you, each player votes for sickness or psychosis. If sickness gets more votes, creatures your opponents control get -2/-2 until end of turn. If psychosis gets more votes or the vote is tied, each opponent discards two cards.
		`,
		},
	}
}
