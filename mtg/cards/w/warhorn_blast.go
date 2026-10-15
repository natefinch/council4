package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WarhornBlast is the card definition for Warhorn Blast.
//
// Type: Instant
// Cost: {4}{W}
//
// Oracle text:
//
//	Creatures you control get +2/+1 until end of turn.
//	Foretell {2}{W} (During your turn, you may pay {2} and exile this card from your hand face down. Cast it on a later turn for its foretell cost.)
var WarhornBlast = newWarhornBlast

func newWarhornBlast() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Warhorn Blast",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Foretell {2}{W}",
					KeywordAbilities: []game.KeywordAbility{
						game.ForetellKeyword{
							Cost: cost.Mana{
								cost.O(2),
								cost.W,
							},
						},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.ApplyContinuous{
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Duration:       game.DurationPermanent,
									Group:          game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
									Layer:          game.LayerPowerToughnessModify,
									PowerDelta:     2,
									ToughnessDelta: 1,
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Creatures you control get +2/+1 until end of turn.
			Foretell {2}{W} (During your turn, you may pay {2} and exile this card from your hand face down. Cast it on a later turn for its foretell cost.)
		`,
		},
	}
}
