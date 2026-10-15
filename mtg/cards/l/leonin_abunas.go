package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LeoninAbunas is the card definition for Leonin Abunas.
//
// Type: Creature — Cat Cleric
// Cost: {3}{W}
//
// Oracle text:
//
//	Artifacts you control have hexproof. (They can't be the targets of spells or abilities your opponents control.)
var LeoninAbunas = newLeoninAbunas

func newLeoninAbunas() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Leonin Abunas",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Cat, types.Cleric},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Artifacts you control have hexproof. (They can't be the targets of spells or abilities your opponents control.)",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Artifact}}),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Hexproof},
						},
					},
				},
			},
			OracleText: `
			Artifacts you control have hexproof. (They can't be the targets of spells or abilities your opponents control.)
		`,
		},
	}
}
