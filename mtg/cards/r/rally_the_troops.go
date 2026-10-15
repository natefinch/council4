package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RallyTheTroops is the card definition for Rally the Troops.
//
// Type: Instant
// Cost: {W}
//
// Oracle text:
//
//	Cast this spell only during the declare attackers step and only if you've been attacked this step.
//	Untap all creatures you control.
var RallyTheTroops = newRallyTheTroops

func newRallyTheTroops() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Rally the Troops",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text:                          "Cast this spell only during the declare attackers step and only if you've been attacked this step.",
					CastOnlyAfterAttackedThisStep: true,
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Untap{
							Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Cast this spell only during the declare attackers step and only if you've been attacked this step.
			Untap all creatures you control.
		`,
		},
	}
}
