package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// EnemyOfTheGuildpact is the card definition for Enemy of the Guildpact.
//
// Type: Creature — Spirit
// Cost: {4}{B}
//
// Oracle text:
//
//	Protection from multicolored
var EnemyOfTheGuildpact = newEnemyOfTheGuildpact

func newEnemyOfTheGuildpact() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Enemy of the Guildpact",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spirit},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.ProtectionFromMulticoloredStaticAbility(),
			},
			OracleText: `
			Protection from multicolored
		`,
		},
	}
}
