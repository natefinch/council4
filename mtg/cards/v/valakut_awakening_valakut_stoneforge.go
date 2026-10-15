package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ValakutAwakening is the card definition for Valakut Awakening // Valakut Stoneforge.
//
// Type: Instant // Land
// Face: Valakut Stoneforge — Land
//
// Oracle text:
//
//	Put any number of cards from your hand on the bottom of your library, then draw that many cards plus one.
var ValakutAwakening = newValakutAwakening

func newValakutAwakening() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Valakut Awakening",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Text: "Put any number of cards from your hand on the bottom of your library, then draw that many cards plus one.",
				Sequence: []game.Instruction{
					{
						Primitive: game.PutHandOnLibraryThenDraw{
							Player:     game.ControllerReference(),
							Bottom:     true,
							DrawOffset: 1,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Put any number of cards from your hand on the bottom of your library, then draw that many cards plus one.
		`,
		},
		Layout: game.LayoutModalDFC,
		Back: opt.Val(game.CardFace{
			Name:  "Valakut Stoneforge",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.R),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			This land enters tapped.
			{T}: Add {R}.
		`,
		}),
	}
}
