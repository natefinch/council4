package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

const breezekeeperOracleID = "138439ad-76c6-4d75-8025-e68b94f68fa3"

var Breezekeeper = cardset.Entry{
	OracleID:     breezekeeperOracleID,
	Name:         "Breezekeeper",
	New:          newBreezekeeper,
	Verification: "mtg/rules/breezekeeper_phasing_test.go: TestBreezekeeperPhasesOutBeforeItsControllerUntaps",
}

func newBreezekeeper() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:      "Breezekeeper",
			ManaCost:  opt.Val(cost.Mana{cost.O(3), cost.U}),
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Djinn},
			Colors:    []color.Color{color.Blue},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			OracleText: "Flying\nPhasing (This phases in or out before you untap during each of your " +
				"untap steps. While it's phased out, it's treated as though it doesn't exist.)",
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.PhasingStaticBody,
			},
		},
	}
}
