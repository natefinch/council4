package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GuardianOfTheGuildpact is the card definition for Guardian of the Guildpact.
//
// Type: Creature — Spirit
// Cost: {3}{W}
//
// Oracle text:
//
//	Protection from monocolored
var GuardianOfTheGuildpact = newGuardianOfTheGuildpact

func newGuardianOfTheGuildpact() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Guardian of the Guildpact",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spirit},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.ProtectionFromMonocoloredStaticAbility(),
			},
			OracleText: `
			Protection from monocolored
		`,
		},
	}
}
