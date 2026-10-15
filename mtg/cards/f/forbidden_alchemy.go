package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ForbiddenAlchemy is the card definition for Forbidden Alchemy.
//
// Type: Instant
// Cost: {2}{U}
//
// Oracle text:
//
//	Look at the top four cards of your library. Put one of them into your hand and the rest into your graveyard.
//	Flashback {6}{B} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
var ForbiddenAlchemy = newForbiddenAlchemy

func newForbiddenAlchemy() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Forbidden Alchemy",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
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
							Look:   game.Fixed(4),
							Take:   game.Fixed(1),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Look at the top four cards of your library. Put one of them into your hand and the rest into your graveyard.
			Flashback {6}{B} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
		`,
		},
	}
}
