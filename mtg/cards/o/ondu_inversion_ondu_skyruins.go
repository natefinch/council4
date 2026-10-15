package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// OnduInversion is the card definition for Ondu Inversion // Ondu Skyruins.
//
// Type: Sorcery // Land
// Face: Ondu Skyruins — Land
//
// Oracle text:
//
//	Destroy all nonland permanents.
var OnduInversion = newOnduInversion

func newOnduInversion() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Ondu Inversion",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
				cost.W,
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Group: game.BattlefieldGroup(game.Selection{ExcludedTypes: []types.Card{types.Land}}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy all nonland permanents.
		`,
		},
		Layout: game.LayoutModalDFC,
		Back: opt.Val(game.CardFace{
			Name:  "Ondu Skyruins",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.W),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			This land enters tapped.
			{T}: Add {W}.
		`,
		}),
	}
}
