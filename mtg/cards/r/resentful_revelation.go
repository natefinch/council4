package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ResentfulRevelation is the card definition for Resentful Revelation.
//
// Type: Sorcery
// Cost: {1}{B}
//
// Oracle text:
//
//	Look at the top three cards of your library. Put one of them into your hand and the rest into your graveyard.
//	Flashback {6}{B} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
var ResentfulRevelation = newResentfulRevelation

func newResentfulRevelation() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Resentful Revelation",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Flashback {6}{B}",
					KeywordAbilities: []game.KeywordAbility{
						game.FlashbackKeyword{
							Cost: cost.Mana{
								cost.O(6),
								cost.B,
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
							Look:   game.Fixed(3),
							Take:   game.Fixed(1),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Look at the top three cards of your library. Put one of them into your hand and the rest into your graveyard.
			Flashback {6}{B} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
		`,
		},
	}
}
