package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SireOfSevenDeaths is the card definition for Sire of Seven Deaths.
//
// Type: Creature — Eldrazi
// Cost: {7}
//
// Oracle text:
//
//	Reach, first strike
//	Vigilance, menace
//	Trample, lifelink
//	Ward—Pay 7 life.
var SireOfSevenDeaths = newSireOfSevenDeaths

func newSireOfSevenDeaths() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Sire of Seven Deaths",
			ManaCost: opt.Val(cost.Mana{
				cost.O(7),
			}),
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Eldrazi},
			Power:     opt.Val(game.PT{Value: 7}),
			Toughness: opt.Val(game.PT{Value: 7}),
			StaticAbilities: []game.StaticAbility{
				game.ReachStaticBody,
				game.FirstStrikeStaticBody,
				game.VigilanceStaticBody,
				game.MenaceStaticBody,
				game.TrampleStaticBody,
				game.LifelinkStaticBody,
				game.WardStaticAbilityWithCosts(cost.Mana{}, []cost.Additional{
					{
						Kind:   cost.AdditionalPayLife,
						Text:   "Pay 7 life",
						Amount: 7,
					},
				}),
			},
			OracleText: `
			Reach, first strike
			Vigilance, menace
			Trample, lifelink
			Ward—Pay 7 life.
		`,
		},
	}
}
