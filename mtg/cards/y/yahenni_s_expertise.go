package y

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// YahenniSExpertise is the card definition for Yahenni's Expertise.
//
// Type: Sorcery
// Cost: {2}{B}{B}
//
// Oracle text:
//
//	All creatures get -3/-3 until end of turn.
//	You may cast a spell with mana value 3 or less from your hand without paying its mana cost.
var YahenniSExpertise = newYahenniSExpertise

func newYahenniSExpertise() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Yahenni's Expertise",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.ApplyContinuous{
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Duration:       game.DurationPermanent,
									Group:          game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
									Layer:          game.LayerPowerToughnessModify,
									PowerDelta:     -3,
									ToughnessDelta: -3,
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
					{
						Primitive: game.CastForFree{
							Player: game.ControllerReference(),
							Selection: game.Selection{ExcludedTypes: []types.Card{types.Land}, ManaValue: opt.Val(compare.Int{
								Op:    compare.LessOrEqual,
								Value: 3,
							})},
							Zone: zone.Hand,
						},
						Optional: true,
					},
				},
			}.Ability()),
			OracleText: `
			All creatures get -3/-3 until end of turn.
			You may cast a spell with mana value 3 or less from your hand without paying its mana cost.
		`,
		},
	}
}
