package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ChillOfForeboding is the card definition for Chill of Foreboding.
//
// Type: Sorcery
// Cost: {2}{U}
//
// Oracle text:
//
//	Each player mills five cards.
//	Flashback {7}{U} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
var ChillOfForeboding = newChillOfForeboding

func newChillOfForeboding() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Chill of Foreboding",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.FlashbackKeyword{Cost: cost.Mana{cost.O(7), cost.U}},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Mill{
							Amount:      game.Fixed(5),
							PlayerGroup: game.AllPlayersReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Each player mills five cards.
			Flashback {7}{U} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
		`,
		},
	}
}
