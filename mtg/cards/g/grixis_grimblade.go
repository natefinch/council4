package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GrixisGrimblade is the card definition for Grixis Grimblade.
//
// Type: Creature — Zombie Warrior
// Cost: {U/R}{B}
//
// Oracle text:
//
//	As long as you control another multicolored permanent, this creature gets +1/+1 and has deathtouch. (Any amount of damage it deals to a creature is enough to destroy that creature.)
var GrixisGrimblade = newGrixisGrimblade

func newGrixisGrimblade() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Grixis Grimblade",
			ManaCost: opt.Val(cost.Mana{
				cost.HybridMana(mana.U, mana.R),
				cost.B,
			}),
			Colors:    []color.Color{color.Black, color.Red, color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie, types.Warrior},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "As long as you control another multicolored permanent, this creature gets +1/+1 and has deathtouch. (Any amount of damage it deals to a creature is enough to destroy that creature.)",
					Condition: opt.Val(game.Condition{
						Text: "As long as you control another multicolored permanent",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{Multicolored: true, ExcludeSource: true},
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerAbility,
							AddKeywords:    []game.Keyword{game.Deathtouch},
						},
					},
				},
			},
			OracleText: `
			As long as you control another multicolored permanent, this creature gets +1/+1 and has deathtouch. (Any amount of damage it deals to a creature is enough to destroy that creature.)
		`,
		},
	}
}
