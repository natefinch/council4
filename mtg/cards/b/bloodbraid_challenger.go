package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// BloodbraidChallenger is the card definition for Bloodbraid Challenger.
//
// Type: Creature — Elf Berserker
// Cost: {3}{R}{G}
//
// Oracle text:
//
//	Cascade
//	Haste
//	Escape—{3}{R}{G}, Exile three other cards from your graveyard. (You may cast this card from your graveyard for its escape cost.)
var BloodbraidChallenger = newBloodbraidChallenger

func newBloodbraidChallenger() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Bloodbraid Challenger",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
				cost.G,
			}),
			Colors:    []color.Color{color.Green, color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elf, types.Berserker},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.CascadeStaticBody,
				game.HasteStaticBody,
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
					ManaCost: opt.Val(cost.Mana{cost.O(3), cost.R, cost.G}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:          cost.AdditionalExile,
							Text:          "Exile three other cards from your graveyard",
							Amount:        3,
							Source:        zone.Graveyard,
							ExcludeSource: true,
						},
					},
					Mechanic: cost.AlternativeMechanicEscape,
				},
			},
			OracleText: `
			Cascade
			Haste
			Escape—{3}{R}{G}, Exile three other cards from your graveyard. (You may cast this card from your graveyard for its escape cost.)
		`,
		},
	}
}
