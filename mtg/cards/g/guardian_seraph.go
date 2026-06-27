package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GuardianSeraph is the card definition for Guardian Seraph.
//
// Type: Creature — Angel
// Cost: {2}{W}{W}
//
// Oracle text:
//
//	Flying
//	If a source an opponent controls would deal damage to you, prevent 1 of that damage.
var GuardianSeraph = newGuardianSeraph()

func newGuardianSeraph() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Guardian Seraph",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Angel},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.DamagePreventionReplacement("If a source an opponent controls would deal damage to you, prevent 1 of that damage.", &game.DamagePreventionSpec{Amount: 1, SourceColors: nil, SourceTypes: nil, SourceControllerOpponent: true}),
			},
			OracleText: `
			Flying
			If a source an opponent controls would deal damage to you, prevent 1 of that damage.
		`,
		},
	}
}
