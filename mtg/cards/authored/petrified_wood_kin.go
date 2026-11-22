package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

var PetrifiedWoodKin = cardset.Entry{
	OracleID: "3bc9649c-4350-4c6a-8441-dc76a5f833af",
	Name:     "Petrified Wood-Kin",
	New:      newPetrifiedWoodKin,
	Verification: "mtg/rules/petrified_wood_kin_test.go: " +
		"TestPetrifiedWoodKinCastsUncounterablyWithBloodthirstAndProtectionFromInstants",
}

func newPetrifiedWoodKin() *game.CardDef {
	const oracleText = "This spell can't be countered.\n" +
		"Bloodthirst X (This creature enters with X +1/+1 counters on it, where X is the damage dealt to your opponents this turn.)\n" +
		"Protection from instants"

	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name:       "Petrified Wood-Kin",
			OracleText: oracleText,
			ManaCost:   opt.Val(cost.Mana{cost.O(6), cost.G}),
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Elemental, types.Warrior},
			Colors:     []color.Color{color.Green},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.CantBeCounteredStaticBody,
				game.ProtectionFromTypesStaticAbility(types.Instant),
			},
			ReplacementAbilities: []game.ReplacementAbility{game.BloodthirstXReplacement("Bloodthirst X")},
		},
	}
}
