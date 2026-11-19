package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

var MerfolkRaiders = cardset.Entry{
	OracleID:     "19edc98c-edc2-47bf-b2da-00f536d7c158",
	Name:         "Merfolk Raiders",
	New:          newMerfolkRaiders,
	Verification: "mtg/rules/phasing_creatures_test.go: TestMerfolkRaidersIslandwalkPreventsBlockingAfterCast",
}

func newMerfolkRaiders() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:      "Merfolk Raiders",
			ManaCost:  opt.Val(cost.Mana{cost.O(1), cost.U}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Merfolk, types.Soldier},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			OracleText: "Islandwalk (This creature can't be blocked as long as defending player controls an Island.)\n" +
				"Phasing (This phases in or out before you untap during each of your untap steps. While it's phased out, it's treated as though it doesn't exist.)",
			StaticAbilities: []game.StaticAbility{
				game.IslandwalkStaticBody,
				game.PhasingStaticBody,
			},
		},
	}
}

var TeferisDrake = cardset.Entry{
	OracleID:     "3605c5ac-3b35-41a5-9266-286fb1d11716",
	Name:         "Teferi's Drake",
	New:          newTeferisDrake,
	Verification: "mtg/rules/phasing_creatures_test.go: TestTeferisDrakeCastsWithFlyingAndPhasing",
}

func newTeferisDrake() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:      "Teferi's Drake",
			ManaCost:  opt.Val(cost.Mana{cost.O(2), cost.U}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Drake},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			OracleText: "Flying\n" +
				"Phasing (This phases in or out before you untap during each of your untap steps. While it's phased out, it's treated as though it doesn't exist.)",
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.PhasingStaticBody,
			},
		},
	}
}

var TolarianDrake = cardset.Entry{
	OracleID:     "d999f2a3-5cb4-4f12-a79a-c5db0a300620",
	Name:         "Tolarian Drake",
	New:          newTolarianDrake,
	Verification: "mtg/rules/phasing_creatures_test.go: TestTolarianDrakeFlyingPreventsGroundBlockAfterCast",
}

func newTolarianDrake() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:      "Tolarian Drake",
			ManaCost:  opt.Val(cost.Mana{cost.O(2), cost.U}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Drake},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 4}),
			OracleText: "Flying\n" +
				"Phasing (This phases in or out before you untap during each of your untap steps. While it's phased out, it's treated as though it doesn't exist.)",
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.PhasingStaticBody,
			},
		},
	}
}

var SandbarCrocodile = cardset.Entry{
	OracleID:     "e11926bf-ca14-4e1e-aca5-9ffedb91868c",
	Name:         "Sandbar Crocodile",
	New:          newSandbarCrocodile,
	Verification: "mtg/rules/phasing_creatures_test.go: TestSandbarCrocodilePhasesOutBeforeItsControllerUntaps",
}

func newSandbarCrocodile() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:      "Sandbar Crocodile",
			ManaCost:  opt.Val(cost.Mana{cost.O(4), cost.U}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Crocodile},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 5}),
			OracleText: "Phasing (This phases in or out before you untap during each of your untap steps. " +
				"While it's phased out, it's treated as though it doesn't exist.)",
			StaticAbilities: []game.StaticAbility{
				game.PhasingStaticBody,
			},
		},
	}
}
