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

// BloodMoon uses the shared CR 305.7 type-layer operation, not a resolving program.
var BloodMoon = cardset.Entry{
	OracleID: "94fac5fe-97d5-4c12-a80c-8efff9d853ae", Name: "Blood Moon", New: newBloodMoon,
}

func newBloodMoon() *game.CardDef {
	const text = "Nonbasic lands are Mountains."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Blood Moon", ManaCost: opt.Val(cost.Mana{cost.O(2), cost.R}),
			Types: []types.Card{types.Enchantment}, Colors: []color.Color{color.Red},
			OracleText: text,
			StaticAbilities: []game.StaticAbility{{
				Text: text, ZoneOfFunction: zone.Battlefield,
				ContinuousEffects: []game.ContinuousEffect{{
					Layer: game.LayerType,
					Group: game.BattlefieldGroup(game.Selection{
						RequiredTypes: []types.Card{types.Land}, ExcludedSupertype: types.Basic,
					}),
					SetBasicLandTypes: []types.Sub{types.Mountain},
				}},
			}},
		},
	}
}
