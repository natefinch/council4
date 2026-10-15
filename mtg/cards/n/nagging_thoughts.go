package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// NaggingThoughts is the card definition for Nagging Thoughts.
//
// Type: Sorcery
// Cost: {1}{U}
//
// Oracle text:
//
//	Look at the top two cards of your library. Put one of them into your hand and the other into your graveyard.
//	Madness {1}{U} (If you discard this card, discard it into exile. When you do, cast it for its madness cost or put it into your graveyard.)
var NaggingThoughts = newNaggingThoughts

func newNaggingThoughts() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Nagging Thoughts",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Madness {1}{U}",
					KeywordAbilities: []game.KeywordAbility{
						game.MadnessKeyword{
							Cost: cost.Mana{
								cost.O(1),
								cost.U,
							},
						},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Dig{
							Player: game.ControllerReference(),
							Look:   game.Fixed(2),
							Take:   game.Fixed(1),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Look at the top two cards of your library. Put one of them into your hand and the other into your graveyard.
			Madness {1}{U} (If you discard this card, discard it into exile. When you do, cast it for its madness cost or put it into your graveyard.)
		`,
		},
	}
}
