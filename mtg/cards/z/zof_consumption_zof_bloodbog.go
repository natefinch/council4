package z

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ZofConsumption is the card definition for Zof Consumption // Zof Bloodbog.
//
// Type: Sorcery // Land
// Face: Zof Bloodbog — Land
//
// Oracle text:
//
//	Each opponent loses 4 life and you gain 4 life.
var ZofConsumption = newZofConsumption

func newZofConsumption() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Zof Consumption",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.LoseLife{
							Amount:      game.Fixed(4),
							PlayerGroup: game.OpponentsReference(),
						},
						PublishResult: game.ResultKey("life-change"),
					},
					{
						Primitive: game.GainLife{
							Amount: game.Fixed(4),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Each opponent loses 4 life and you gain 4 life.
		`,
		},
		Layout: game.LayoutModalDFC,
		Back: opt.Val(game.CardFace{
			Name:  "Zof Bloodbog",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.B),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			This land enters tapped.
			{T}: Add {B}.
		`,
		}),
	}
}
