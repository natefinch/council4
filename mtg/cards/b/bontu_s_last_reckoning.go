package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BontuSLastReckoning is the card definition for Bontu's Last Reckoning.
//
// Type: Sorcery
// Cost: {1}{B}{B}
//
// Oracle text:
//
//	Destroy all creatures. Lands you control don't untap during your next untap step.
var BontuSLastReckoning = newBontuSLastReckoning()

func newBontuSLastReckoning() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Bontu's Last Reckoning",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
						},
					},
					{
						Primitive: game.SkipNextUntap{
							Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Land}, Controller: game.ControllerYou}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy all creatures. Lands you control don't untap during your next untap step.
		`,
		},
	}
}
