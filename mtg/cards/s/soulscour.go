package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Soulscour is the card definition for Soulscour.
//
// Type: Sorcery
// Cost: {7}{W}{W}{W}
//
// Oracle text:
//
//	Destroy all nonartifact permanents.
var Soulscour = newSoulscour

func newSoulscour() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Soulscour",
			ManaCost: opt.Val(cost.Mana{
				cost.O(7),
				cost.W,
				cost.W,
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Group: game.BattlefieldGroup(game.Selection{ExcludedTypes: []types.Card{types.Artifact}}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy all nonartifact permanents.
		`,
		},
	}
}
