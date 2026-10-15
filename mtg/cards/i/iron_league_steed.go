package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// IronLeagueSteed is the card definition for Iron League Steed.
//
// Type: Artifact Creature — Construct
// Cost: {4}
//
// Oracle text:
//
//	Haste
//	Fabricate 1 (When this creature enters, put a +1/+1 counter on it or create a 1/1 colorless Servo artifact creature token.)
var IronLeagueSteed = newIronLeagueSteed

func newIronLeagueSteed() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Iron League Steed",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Construct},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.HasteStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.FabricateTriggeredAbility(1),
			},
			OracleText: `
			Haste
			Fabricate 1 (When this creature enters, put a +1/+1 counter on it or create a 1/1 colorless Servo artifact creature token.)
		`,
		},
	}
}
