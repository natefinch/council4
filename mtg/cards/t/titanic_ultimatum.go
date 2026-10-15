package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TitanicUltimatum is the card definition for Titanic Ultimatum.
//
// Type: Sorcery
// Cost: {R}{R}{G}{G}{G}{W}{W}
//
// Oracle text:
//
//	Until end of turn, creatures you control get +5/+5 and gain first strike, trample, and lifelink.
var TitanicUltimatum = newTitanicUltimatum

func newTitanicUltimatum() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Titanic Ultimatum",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
				cost.R,
				cost.G,
				cost.G,
				cost.G,
				cost.W,
				cost.W,
			}),
			Colors: []color.Color{color.Green, color.Red, color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.ApplyContinuous{
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Duration:       game.DurationPermanent,
									Group:          game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
									Layer:          game.LayerPowerToughnessModify,
									PowerDelta:     5,
									ToughnessDelta: 5,
								},
								game.ContinuousEffect{
									Duration:    game.DurationPermanent,
									Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
									Layer:       game.LayerAbility,
									AddKeywords: []game.Keyword{game.FirstStrike, game.Trample, game.Lifelink},
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Until end of turn, creatures you control get +5/+5 and gain first strike, trample, and lifelink.
		`,
		},
	}
}
