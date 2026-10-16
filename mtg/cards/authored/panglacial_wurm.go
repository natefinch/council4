package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// PanglacialWurm has a library-only permission, not a resolving spell program.
var PanglacialWurm = cardset.Entry{
	OracleID: "49d4aef7-d2d1-4f66-bc2a-79ec8cc96565", Name: "Panglacial Wurm", New: newPanglacialWurm,
	Verification: "mtg/rules/panglacial_wurm_test.go; WotC 2006-07-15; CR 2026-09-25 117.2e, 601.2, 605.3, 701.23, 733.1; #3286",
}

func newPanglacialWurm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Panglacial Wurm", OracleText: "Trample\nWhile you're searching your library, you may cast this card from your library.",
			ManaCost: opt.Val(cost.Mana{cost.O(5), cost.G, cost.G}),
			Types:    []types.Card{types.Creature}, Subtypes: []types.Sub{types.Wurm},
			Colors: []color.Color{color.Green}, Power: opt.Val(game.PT{Value: 9}), Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{game.TrampleStaticBody, {
				Text:           "While you're searching your library, you may cast this card from your library.",
				ZoneOfFunction: zone.Library, CastWhileSearchingOwnLibrary: true,
			}},
		},
	}
}
