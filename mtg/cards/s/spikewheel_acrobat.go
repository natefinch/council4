package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SpikewheelAcrobat is the card definition for Spikewheel Acrobat.
//
// Type: Creature — Human Rogue
// Cost: {3}{R}
//
// Oracle text:
//
//	Spectacle {2}{R} (You may cast this spell for its spectacle cost rather than its mana cost if an opponent lost life this turn.)
var SpikewheelAcrobat = newSpikewheelAcrobat

func newSpikewheelAcrobat() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Spikewheel Acrobat",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Rogue},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 2}),
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label:     "Spectacle",
					ManaCost:  opt.Val(cost.Mana{cost.O(2), cost.R}),
					Condition: cost.AlternativeConditionOpponentLostLifeThisTurn,
				},
			},
			OracleText: `
			Spectacle {2}{R} (You may cast this spell for its spectacle cost rather than its mana cost if an opponent lost life this turn.)
		`,
		},
	}
}
