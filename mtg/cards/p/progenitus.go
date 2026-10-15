package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Progenitus is the card definition for Progenitus.
//
// Type: Legendary Creature — Hydra Avatar
// Cost: {W}{W}{U}{U}{B}{B}{R}{R}{G}{G}
//
// Oracle text:
//
//	Protection from everything
//	If Progenitus would be put into a graveyard from anywhere, reveal Progenitus and shuffle it into its owner's library instead.
var Progenitus = newProgenitus

func newProgenitus() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Black, color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Progenitus",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
				cost.W,
				cost.U,
				cost.U,
				cost.B,
				cost.B,
				cost.R,
				cost.R,
				cost.G,
				cost.G,
			}),
			Colors:     []color.Color{color.Black, color.Green, color.Red, color.Blue, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Hydra, types.Avatar},
			Power:      opt.Val(game.PT{Value: 10}),
			Toughness:  opt.Val(game.PT{Value: 10}),
			StaticAbilities: []game.StaticAbility{
				game.ProtectionFromEverythingStaticAbility(),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.ReplacementAbility{Text: "If Progenitus would be put into a graveyard from anywhere, reveal Progenitus and shuffle it into its owner's library instead.", Replacement: game.ReplacementEffect{
					MatchEvent:         game.EventZoneChanged,
					MatchToZone:        true,
					ToZone:             zone.Graveyard,
					ReplaceToZone:      zone.Library,
					Duration:           game.DurationPermanent,
					ShuffleIntoLibrary: true,
					RevealSource:       true,
				}},
			},
			OracleText: `
			Protection from everything
			If Progenitus would be put into a graveyard from anywhere, reveal Progenitus and shuffle it into its owner's library instead.
		`,
		},
	}
}
