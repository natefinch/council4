package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ArcaneInfusion is the card definition for Arcane Infusion.
//
// Type: Instant
// Cost: {U}{R}
//
// Oracle text:
//
//	Look at the top four cards of your library. You may reveal an instant or sorcery card from among them and put it into your hand. Put the rest on the bottom of your library in a random order.
//	Flashback {3}{U}{R} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
var ArcaneInfusion = newArcaneInfusion

func newArcaneInfusion() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red),
		CardFace: game.CardFace{
			Name: "Arcane Infusion",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
				cost.R,
			}),
			Colors: []color.Color{color.Red, color.Blue},
			Types:  []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.FlashbackKeyword{Cost: cost.Mana{cost.O(3), cost.U, cost.R}},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Dig{
							Player:    game.ControllerReference(),
							Look:      game.Fixed(4),
							Take:      game.Fixed(1),
							Remainder: game.DigRemainderLibraryBottom,
							Filter:    opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Instant, types.Sorcery}}),
							TakeUpTo:  true,
							Reveal:    true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Look at the top four cards of your library. You may reveal an instant or sorcery card from among them and put it into your hand. Put the rest on the bottom of your library in a random order.
			Flashback {3}{U}{R} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
		`,
		},
	}
}
