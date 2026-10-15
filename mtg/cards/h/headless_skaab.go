package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// HeadlessSkaab is the card definition for Headless Skaab.
//
// Type: Creature — Zombie Warrior
// Cost: {2}{U}
//
// Oracle text:
//
//	As an additional cost to cast this spell, exile a creature card from your graveyard.
//	This creature enters tapped.
var HeadlessSkaab = newHeadlessSkaab

func newHeadlessSkaab() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Headless Skaab",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie, types.Warrior},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 6}),
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This creature enters tapped."),
			},
			AdditionalCosts: []cost.Additional{
				{
					Kind:          cost.AdditionalExile,
					Text:          "exile a creature card from your graveyard",
					Amount:        1,
					Source:        zone.Graveyard,
					MatchCardType: true,
					CardType:      types.Creature,
				},
			},
			OracleText: `
			As an additional cost to cast this spell, exile a creature card from your graveyard.
			This creature enters tapped.
		`,
		},
	}
}
