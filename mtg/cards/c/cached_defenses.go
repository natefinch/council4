package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CachedDefenses is the card definition for Cached Defenses.
//
// Type: Sorcery
// Cost: {2}{G}
//
// Oracle text:
//
//	Bolster 3. (Choose a creature with the least toughness among creatures you control and put three +1/+1 counters on it.)
var CachedDefenses = newCachedDefenses

func newCachedDefenses() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Cached Defenses",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Bolster{
							Amount: game.Fixed(3),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Bolster 3. (Choose a creature with the least toughness among creatures you control and put three +1/+1 counters on it.)
		`,
		},
	}
}
