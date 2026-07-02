package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Falter is the card definition for Falter.
//
// Type: Instant
// Cost: {1}{R}
//
// Oracle text:
//
//	Creatures without flying can't block this turn.
var Falter = newFalter()

func newFalter() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Falter",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.ApplyRule{
							RuleEffects: []game.RuleEffect{
								game.RuleEffect{
									Kind:              game.RuleEffectCantBlock,
									PermanentTypes:    []types.Card{types.Creature},
									AffectedSelection: game.Selection{ExcludedKeyword: game.Flying},
								},
							},
							Duration: game.DurationThisTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Creatures without flying can't block this turn.
		`,
		},
	}
}
