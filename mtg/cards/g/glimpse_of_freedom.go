package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GlimpseOfFreedom is the card definition for Glimpse of Freedom.
//
// Type: Instant
// Cost: {1}{U}
//
// Oracle text:
//
//	Draw a card.
//	Escape—{2}{U}, Exile five other cards from your graveyard. (You may cast this card from your graveyard for its escape cost.)
var GlimpseOfFreedom = newGlimpseOfFreedom

func newGlimpseOfFreedom() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Glimpse of Freedom",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.SimpleKeyword{
							Kind: game.Escape,
						},
					},
				},
			},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label:    "Escape",
					ManaCost: opt.Val(cost.Mana{cost.O(2), cost.U}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:          cost.AdditionalExile,
							Text:          "Exile five other cards from your graveyard",
							Amount:        5,
							Source:        zone.Graveyard,
							ExcludeSource: true,
						},
					},
					Mechanic: cost.AlternativeMechanicEscape,
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Draw{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Draw a card.
			Escape—{2}{U}, Exile five other cards from your graveyard. (You may cast this card from your graveyard for its escape cost.)
		`,
		},
	}
}
