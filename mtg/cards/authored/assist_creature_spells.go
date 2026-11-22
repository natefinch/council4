package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

const (
	chargingBinoxOracleID = "0f10f9d1-f9f9-4356-9208-0f21f44bfb97"
	magmaHellionOracleID  = "88f6575d-a6cb-4b2c-90dd-c6239f3c6e4d"
)

var (
	ChargingBinox = cardset.Entry{
		OracleID: chargingBinoxOracleID,
		Name:     "Charging Binox",
		New:      newChargingBinox,
		Verification: "mtg/rules/assist_creature_spells_test.go: " +
			"TestChargingBinoxAssistAndTrample",
	}
	MagmaHellion = cardset.Entry{
		OracleID: magmaHellionOracleID,
		Name:     "Magma Hellion",
		New:      newMagmaHellion,
		Verification: "mtg/rules/assist_creature_spells_test.go: " +
			"TestMagmaHellionAssistTrampleAndHaste",
	}
)

func newChargingBinox() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name:       "Charging Binox",
			ManaCost:   opt.Val(cost.Mana{cost.O(7), cost.G}),
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Beast},
			Colors:     []color.Color{color.Green},
			Power:      opt.Val(game.PT{Value: 7}),
			Toughness:  opt.Val(game.PT{Value: 5}),
			OracleText: "Assist (Another player can pay up to {7} of this spell's cost.)\nTrample",
			StaticAbilities: []game.StaticAbility{
				game.AssistStaticBody,
				game.TrampleStaticBody,
			},
		},
	}
}

func newMagmaHellion() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:       "Magma Hellion",
			ManaCost:   opt.Val(cost.Mana{cost.O(6), cost.R}),
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Hellion},
			Colors:     []color.Color{color.Red},
			Power:      opt.Val(game.PT{Value: 5}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			OracleText: "Assist (Another player can pay up to {6} of this spell's cost.)\nTrample, haste",
			StaticAbilities: []game.StaticAbility{
				game.AssistStaticBody,
				game.TrampleStaticBody,
				game.HasteStaticBody,
			},
		},
	}
}
