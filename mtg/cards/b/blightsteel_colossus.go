package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// BlightsteelColossus is the card definition for Blightsteel Colossus.
//
// Type: Artifact Creature — Phyrexian Golem
// Cost: {12}
//
// Oracle text:
//
//	Trample, infect, indestructible
//	If Blightsteel Colossus would be put into a graveyard from anywhere, reveal Blightsteel Colossus and shuffle it into its owner's library instead.
var BlightsteelColossus = newBlightsteelColossus

func newBlightsteelColossus() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Blightsteel Colossus",
			ManaCost: opt.Val(cost.Mana{
				cost.O(12),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian, types.Golem},
			Power:     opt.Val(game.PT{Value: 11}),
			Toughness: opt.Val(game.PT{Value: 11}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
				game.InfectStaticBody,
				game.IndestructibleStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.ReplacementAbility{Text: "If Blightsteel Colossus would be put into a graveyard from anywhere, reveal Blightsteel Colossus and shuffle it into its owner's library instead.", Replacement: game.ReplacementEffect{
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
			Trample, infect, indestructible
			If Blightsteel Colossus would be put into a graveyard from anywhere, reveal Blightsteel Colossus and shuffle it into its owner's library instead.
		`,
		},
	}
}
