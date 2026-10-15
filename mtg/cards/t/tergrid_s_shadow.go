package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TergridSShadow is the card definition for Tergrid's Shadow.
//
// Type: Instant
// Cost: {3}{B}{B}
//
// Oracle text:
//
//	Each player sacrifices two creatures.
//	Foretell {2}{B}{B} (During your turn, you may pay {2} and exile this card from your hand face down. Cast it on a later turn for its foretell cost.)
var TergridSShadow = newTergridSShadow

func newTergridSShadow() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Tergrid's Shadow",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Foretell {2}{B}{B}",
					KeywordAbilities: []game.KeywordAbility{
						game.ForetellKeyword{
							Cost: cost.Mana{
								cost.O(2),
								cost.B,
								cost.B,
							},
						},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.SacrificePermanents{
							PlayerGroup: game.AllPlayersReference(),
							Amount:      game.Fixed(2),
							Selection:   game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
				},
			}.Ability()),
			OracleText: `
			Each player sacrifices two creatures.
			Foretell {2}{B}{B} (During your turn, you may pay {2} and exile this card from your hand face down. Cast it on a later turn for its foretell cost.)
		`,
		},
	}
}
