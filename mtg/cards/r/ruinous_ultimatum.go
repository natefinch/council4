package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RuinousUltimatum is the card definition for Ruinous Ultimatum.
//
// Type: Sorcery
// Cost: {R}{R}{W}{W}{W}{B}{B}
//
// Oracle text:
//
//	Destroy all nonland permanents your opponents control.
var RuinousUltimatum = newRuinousUltimatum

func newRuinousUltimatum() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Ruinous Ultimatum",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
				cost.R,
				cost.W,
				cost.W,
				cost.W,
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black, color.Red, color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Group: game.BattlefieldGroup(game.Selection{ExcludedTypes: []types.Card{types.Land}, Controller: game.ControllerOpponent}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy all nonland permanents your opponents control.
		`,
		},
	}
}
