package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BlindingLight is the card definition for Blinding Light.
//
// Type: Sorcery
// Cost: {2}{W}
//
// Oracle text:
//
//	Tap all nonwhite creatures.
var BlindingLight = newBlindingLight

func newBlindingLight() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Blinding Light",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Tap{
							Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, ExcludedColors: []color.Color{color.White}}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Tap all nonwhite creatures.
		`,
		},
	}
}
