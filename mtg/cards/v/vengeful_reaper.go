package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// VengefulReaper is the card definition for Vengeful Reaper.
//
// Type: Creature — Angel Cleric
// Cost: {3}{B}
//
// Oracle text:
//
//	Flying, deathtouch, haste
//	Foretell {1}{B} (During your turn, you may pay {2} and exile this card from your hand face down. Cast it on a later turn for its foretell cost.)
var VengefulReaper = newVengefulReaper

func newVengefulReaper() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Vengeful Reaper",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Angel, types.Cleric},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.DeathtouchStaticBody,
				game.HasteStaticBody,
				game.StaticAbility{
					Text: "Foretell {1}{B}",
					KeywordAbilities: []game.KeywordAbility{
						game.ForetellKeyword{
							Cost: cost.Mana{
								cost.O(1),
								cost.B,
							},
						},
					},
				},
			},
			OracleText: `
			Flying, deathtouch, haste
			Foretell {1}{B} (During your turn, you may pay {2} and exile this card from your hand face down. Cast it on a later turn for its foretell cost.)
		`,
		},
	}
}
