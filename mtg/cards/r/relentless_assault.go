package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RelentlessAssault is the card definition for Relentless Assault.
//
// Type: Sorcery
// Cost: {2}{R}{R}
//
// Oracle text:
//
//	Untap all creatures that attacked this turn. After this main phase, there is an additional combat phase followed by an additional main phase.
var RelentlessAssault = newRelentlessAssault

func newRelentlessAssault() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Relentless Assault",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Untap{
							Group: game.AttackedThisTurnGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
						},
					},
					{
						Primitive: game.AddExtraPhases{
							Combat: true,
							Main:   true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Untap all creatures that attacked this turn. After this main phase, there is an additional combat phase followed by an additional main phase.
		`,
		},
	}
}
