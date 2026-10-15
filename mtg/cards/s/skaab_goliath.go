package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SkaabGoliath is the card definition for Skaab Goliath.
//
// Type: Creature — Zombie Giant
// Cost: {5}{U}
//
// Oracle text:
//
//	As an additional cost to cast this spell, exile two creature cards from your graveyard.
//	Trample
var SkaabGoliath = newSkaabGoliath

func newSkaabGoliath() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Skaab Goliath",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie, types.Giant},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 9}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
			},
			AdditionalCosts: []cost.Additional{
				{
					Kind:          cost.AdditionalExile,
					Text:          "exile two creature cards from your graveyard",
					Amount:        2,
					Source:        zone.Graveyard,
					MatchCardType: true,
					CardType:      types.Creature,
				},
			},
			OracleText: `
			As an additional cost to cast this spell, exile two creature cards from your graveyard.
			Trample
		`,
		},
	}
}
