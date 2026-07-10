package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ScreamreachBrawler is the card definition for Screamreach Brawler.
//
// Type: Creature — Orc Berserker
// Cost: {2}{R}
//
// Oracle text:
//
//	Dash {1}{R} (You may cast this spell for its dash cost. If you do, it gains haste, and it's returned from the battlefield to its owner's hand at the beginning of the next end step.)
var ScreamreachBrawler = newScreamreachBrawler

func newScreamreachBrawler() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Screamreach Brawler",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Orc, types.Berserker},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.DashTriggeredAbility(),
			},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label:    "Dash",
					ManaCost: opt.Val(cost.Mana{cost.O(1), cost.R}),
					Mechanic: cost.AlternativeMechanicDash,
				},
			},
			OracleText: `
			Dash {1}{R} (You may cast this spell for its dash cost. If you do, it gains haste, and it's returned from the battlefield to its owner's hand at the beginning of the next end step.)
		`,
		},
	}
}
